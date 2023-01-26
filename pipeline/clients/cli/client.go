package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/syncutil"
	"github.com/grafana/scribe/wrappers"
	"github.com/sirupsen/logrus"
)

// The Client is used when interacting with a scribe pipeline using the scribe CLI. It is used to run only one step.
// The CLI client simply runs the anonymous function defined in the step.
type Client struct {
	Opts  clients.CommonOpts
	Log   *logrus.Logger
	State *StateWrapper
}

func New(ctx context.Context, opts clients.CommonOpts) (pipeline.Client, error) {
	if opts.Args.Step == nil || *opts.Args.Step == 0 {
		return nil, errors.New("--step argument can not be empty or 0 when using the CLI client")
	}

	return &Client{
		Opts: opts,
		Log:  opts.Log,
		State: NewStateWrapper(
			state.ReaderWithLogs(opts.Log, state.NewArgMapReader(opts.Args.ArgMap)),
			&StateHandler{},
		),
	}, nil
}

// PipelineWalkFunc walks through the pipelines that the collection provides. Each pipeline is a pipeline of steps, so each will walk through the list of steps using the StepWalkFunc.
func (c *Client) HandlePipeline(ctx context.Context, p pipeline.Pipeline) error {
	var (
		wg = syncutil.NewStepWaitGroup()
	)

	for _, node := range p.Graph.Nodes {
		// Skip the root step that's always present on every pipeline.
		if node.ID == 0 {
			continue
		}
		log := c.Opts.Log
		logWrapper := &wrappers.LogWrapper{
			Opts: c.Opts,
			Log:  log.WithField("step", node.Value.Name),
		}
		traceWrapper := &wrappers.TraceWrapper{
			Opts:   c.Opts,
			Tracer: c.Opts.Tracer,
		}

		step := logWrapper.WrapStep(node.Value)
		step = traceWrapper.WrapStep(step)

		// Otherwise, add this pipeline to the set that needs to complete before moving on to the next set of pipelines.
		wg.Add(step, pipeline.ActionOpts{
			Path:    c.Opts.Args.Path,
			State:   c.State,
			Tracer:  c.Opts.Tracer,
			Version: c.Opts.Version,
			Logger:  log,
		})
	}

	if err := wg.Wait(ctx); err != nil {
		return err
	}

	if err := json.NewEncoder(os.Stdout).Encode(c.State.data); err != nil {
		return fmt.Errorf("error encoding JSON for CLI client state updates: %w", err)
	}

	return nil
}

func (c *Client) Validate(step pipeline.Step) error {
	return nil
}

func (c *Client) HandleEvents(events []pipeline.Event) error {
	return nil
}

func (c *Client) Done(ctx context.Context, w *pipeline.Collection) error {
	for _, node := range w.Graph.Nodes {
		// Skip the root node because there's always a root node that just exists as a starting point.
		if node.ID == 0 {
			continue
		}
		pipeline := node.Value
		// Not counting the root pipeline, there should really only be 1 pipeline here with 1 step since we've filtered by step ID.
		if err := c.HandlePipeline(ctx, pipeline); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) prepopulateState(ctx context.Context, s state.Handler) error {
	log := c.Log
	for k, v := range KnownValues {
		exists, err := s.Exists(ctx, k)
		if err != nil {
			// Even if we encounter an error, we still want to attempt to set the state.
			// One error that could happen here is if the state is empty.
			log.WithError(err).Debugln("Failed to read state")
		}

		if !exists {
			log.Debugln("State not found for", k.Key, "preopulating value")
			if err := v(ctx, s); err != nil {
				log.WithError(err).Debugln("Failed to pre-populate state for argument", k.Key)
			}
		}
	}

	return nil
}
