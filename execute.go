package scribe

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/plog"
	"github.com/grafana/scribe/state"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	"golang.org/x/exp/slices"
)

var ArgDefaults = map[string]func(context.Context) *exec.Cmd{
	pipeline.ArgumentCommitRef.Key: func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	},
	pipeline.ArgumentCommitSHA.Key: func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	},
	pipeline.ArgumentRemoteURL.Key: func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	},
	pipeline.ArgumentBranch.Key: func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	},
	pipeline.ArgumentTagName.Key: func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "git", "describe", "--tags")
	},
}

// executeFunc is shared between the Scribe and ScribeMulti types.
// Because the behavior of processing the pipeline is essentially the same, and they should behave the same in perpituity,
// these functions ensure that they at least behave consistently.
type executeFunc func(context.Context, *pipeline.Collection) error

func executeWithTracing(tracer opentracing.Tracer, ef executeFunc) executeFunc {
	return func(ctx context.Context, collection *pipeline.Collection) error {
		span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "scribe")
		defer span.Finish()
		err := ef(ctx, collection)
		if v, ok := tracer.(*jaeger.Tracer); ok {
			v.Close()
		}

		return err
	}
}

func executeWithLogging(log logrus.FieldLogger, ef executeFunc) executeFunc {
	return func(ctx context.Context, collection *pipeline.Collection) error {
		err := ef(ctx, collection)
		if err != nil {
			if errors.Is(err, ErrorCancelled) {
				log.WithFields(logrus.Fields{
					"status":       "cancelled",
					"completed_at": time.Now().Unix(),
				}).WithError(err).Infoln("execution completed")
			} else {
				log.WithFields(logrus.Fields{
					"status":       "error",
					"completed_at": time.Now().Unix(),
				}).WithError(err).Infoln("execution completed")
			}

			return err
		}

		log.WithFields(logrus.Fields{
			"status":       "success",
			"completed_at": time.Now().Unix(),
		}).Info("execution completed")

		return nil
	}
}

func executeWithSteps(
	args *args.PipelineArgs,
	name string,
	n *counter,
	ef executeFunc,
) executeFunc {
	return func(ctx context.Context, collection *pipeline.Collection) error {
		// If the user has specified a specific step, then cut the "Collection" to only include that step
		if args.Step != nil {
			step, err := collection.ByID(ctx, *args.Step)
			if err != nil {
				return fmt.Errorf("could not find step with id '%d'. Error: %w", *args.Step, err)
			}
			c, err := pipeline.NewCollectionWithSteps(name, step)
			if err != nil {
				return err
			}
			collection = c
		}

		return ef(ctx, collection)
	}
}

func executeWithPipelines(
	args *args.PipelineArgs,
	name string,
	n *counter,
	ef executeFunc,
) executeFunc {
	return func(ctx context.Context, collection *pipeline.Collection) error {
		// If the user has specified specific pipelines, then cut the "Collection" to only include those pipelines.
		if len(args.PipelineName) != 0 {
			pipelines, err := collection.PipelinesByName(ctx, args.PipelineName)
			if err != nil {
				return fmt.Errorf("could not find any pipelines that match '%v'. Error: %w", args.PipelineName, err)
			}
			c := pipeline.NewCollection()
			if err := c.AddPipelines(pipelines...); err != nil {
				return err
			}
			collection = c
		}

		return ef(ctx, collection)
	}
}

// executeWithEvent uses the provided '-event' argument which will populate the state with data that represents the event.
func executeWithEvent(
	args *args.PipelineArgs,
	opts clients.CommonOpts,
	ef executeFunc,
) executeFunc {
	log := opts.Log
	return func(ctx context.Context, collection *pipeline.Collection) error {
		type event struct {
			Args     []state.Argument
			Pipeline int64
		}

		var (
			pipelineList = map[string]event{}
		)

		for _, p := range collection.Graph.Nodes {
			// For every pipeline, set the arguments that each event requires into the pipeline.
			// By default assume the user has selected the git-commit event
			pipelineList["git-commit"] = event{
				Args:     pipeline.GitCommitEventArgs,
				Pipeline: p.ID,
			}

			// However, still add every event found in the pipelines. This gives us a list of possible events that the user could have selected which we can present to them.
			for _, e := range p.Value.Events {
				// Nailvely add each event to the list. It doesn't matter if we overwrite what's already there because event name collisions shouldn't happen.
				pipelineList[e.Name] = event{
					Args:     e.Provides,
					Pipeline: p.ID,
				}
			}
		}

		if len(pipelineList) == 0 {
			return ef(ctx, collection)
		}

		e := opts.Args.Event
		// If the user has not provided an event argument, then set a default and warn them.
		if e == "" {
			log.Warnln("No event was selected; assuming event is 'git-commit'")
			keys := make([]string, len(pipelineList))
			i := 0
			for k := range pipelineList {
				keys[i] = "'" + k + "'"
				i++
			}

			log.Warnln("Other possible events for this program are:", strings.Join(keys, " "))
			e = "git-commit"
		}

		pipelines, err := collection.PipelinesByEvent(ctx, e)
		if err != nil {
			return fmt.Errorf("error finding pipeline by event '%s': '%w'", e, err)
		}

		c := pipeline.NewCollection()
		if err := c.AddPipelines(pipelines...); err != nil {
			return err
		}

		return ef(ctx, c)
	}
}

func executeWithSignals(
	ef executeFunc,
) executeFunc {
	return func(ctx context.Context, collection *pipeline.Collection) error {
		ctx, cancel := signal.NotifyContext(ctx, os.Interrupt,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		defer cancel()
		return ef(ctx, collection)
	}
}

// LocalClients define modes that are intended to run a pipeline "locally".
// These local clients will do things like filter the pipeline based on the selected event with the '-e' flag.
var LocalModes = []string{"dagger"}

// Execute runs the provided executeFunc with the appropriate wrappers.
// All of the arguments are for populating the wrappers.
func execute(ctx context.Context, collection *pipeline.Collection, name string, opts clients.CommonOpts, n *counter, ef executeFunc) error {
	logger := opts.Log.WithFields(plog.Combine(plog.TracingFields(ctx), plog.PipelineFields(opts)))

	// Wrap with signals watching. If the user submits a SIGTERM/SIGINT/SIGKILL, this function will catch it and return an error.
	wrapped := executeWithSignals(ef)

	// If the user supplies a --event or -e argument, check the arguments for the event and reduce the collection
	// However, we only want to do this type of filtering when we're running locally using the dagger mode.
	if slices.Contains(LocalModes, opts.Args.Client) {
		wrapped = executeWithEvent(opts.Args, opts, wrapped)
	}

	//// If the user supplies a --step argument, reduce the collection
	wrapped = executeWithSteps(opts.Args, name, n, wrapped)

	//// If the user supplies a --pipeline or -p argument, reduce the collection
	wrapped = executeWithPipelines(opts.Args, name, n, wrapped)

	// Add a root tracing span to the context, and end the span when the executeFunc is done.
	wrapped = executeWithTracing(opts.Tracer, wrapped)

	// Add structured logging when the pipeline execution starts and ends.
	wrapped = executeWithLogging(logger, wrapped)

	if err := wrapped(ctx, collection); err != nil {
		return err
	}

	return nil
}
