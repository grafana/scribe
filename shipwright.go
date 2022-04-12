// Package shipwright provides the primary library / client functions, types, and methods for creating Shipwright pipelines.
package shipwright

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/plog"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

const DefaultPipelineID int64 = 1

// Shipwright is the client that is used in every pipeline to declare the steps that make up a pipeline.
// The Shipwright type is not thread safe. Running any of the functions from this type concurrently may have unexpected results.
type Shipwright[T pipeline.StepContent] struct {
	Client     pipeline.Client
	Collection *pipeline.Collection
	Events     []pipeline.Event

	pipeline.Configurer

	// Opts are the options that are provided to the pipeline from outside sources. This includes mostly command-line arguments and environment variables
	Opts pipeline.CommonOpts
	Log  *logrus.Logger

	// n tracks the ID of a step so that the "shipwright -step=" argument will function independently of the client implementation
	// It ensures that the 11th step in a Drone generated pipeline is also the 11th step in a CLI pipeline
	n        int64
	pipeline int64
	Version  string

	prev          []pipeline.Step[pipeline.StepList]
	prevPipelines []pipeline.Step[pipeline.Pipeline]
}

// Pipeline returns the current Pipeline ID used in the collection.
func (s *Shipwright[T]) Pipeline() int64 {
	return s.pipeline
}

// When allows users to define when this pipeline is executed, especially in the remote environment.
func (s *Shipwright[T]) When(event ...pipeline.Event) {
	s.Events = event
}

func (s *Shipwright[T]) newList(steps ...pipeline.Step[pipeline.Action]) pipeline.Step[pipeline.StepList] {
	list := pipeline.Step[pipeline.StepList]{
		Serial:  s.n,
		Content: steps,
	}

	s.n++

	return list
}

// Background allows users to define steps that run in the background. In some environments this is referred to as a "Service" or "Background service".
// In many scenarios, users would like to simply use a docker image with the default command. In order to accomplish that, simply provide a step without an action.
func (s *Shipwright[T]) Background(steps ...pipeline.Step[pipeline.Action]) {
	if err := s.validateSteps(steps...); err != nil {
		s.Log.Fatalln(err)
	}

	st := s.setup(any(steps).([]pipeline.Step[T])...)
	list := s.newList(any(st).([]pipeline.Step[pipeline.Action])...)

	if err := s.Collection.AddSteps(s.pipeline, list); err != nil {
		s.Log.Fatalln(err)
	}
}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the pipeline execution until all of the steps provided (step) have completed sequentially.
func (s *Shipwright[T]) Run(steps ...pipeline.Step[T]) {
	steps = s.setup(steps...)

	switch x := any(steps).(type) {
	case []pipeline.Step[pipeline.Action]:
		if err := s.runSteps(x...); err != nil {
			s.Log.Fatalln(err)
		}
	case []pipeline.Step[pipeline.Pipeline]:
		if err := s.runPipelines(x...); err != nil {
			s.Log.Fatalln(err)
		}
	}
}

func (s *Shipwright[T]) runSteps(steps ...pipeline.Step[pipeline.Action]) error {
	if err := s.validateSteps(steps...); err != nil {
		return err
	}

	prev := s.prev

	for _, v := range steps {
		list := s.newList(v)
		list.Dependencies = prev

		if err := s.Collection.AddSteps(s.pipeline, list); err != nil {
			return fmt.Errorf("error adding step '%d' to collection. error: %w", list.Serial, err)
		}

		prev = []pipeline.Step[pipeline.StepList]{list}
	}

	s.prev = prev

	return nil
}

// runPipeliens adds the list of pipelines to the collection. Pipelines are essentially branches in the graph.
// The pipelines provided run one after another.
func (s *Shipwright[T]) runPipelines(pipelines ...pipeline.Step[pipeline.Pipeline]) error {
	prev := s.prevPipelines

	for _, v := range pipelines {
		v.Dependencies = prev
		if err := s.Collection.AddPipelines(v); err != nil {
			return fmt.Errorf("error adding pipeline '%d' to collection. error: %w", v.Serial, err)
		}

		prev = []pipeline.Step[pipeline.Pipeline]{v}
	}

	return nil
}

// Parallel will run the listed steps at the same time.
// This function blocks the pipeline execution until all of the steps have completed.
func (s *Shipwright[T]) Parallel(steps ...pipeline.Step[T]) {
	steps = s.setup(steps...)

	switch x := any(steps).(type) {
	case []pipeline.Step[pipeline.Action]:
		if err := s.parallelSteps(x...); err != nil {
			s.Log.Fatalln(err)
		}
	case []pipeline.Step[pipeline.Pipeline]:
		if err := s.parallelPipelines(x...); err != nil {
			s.Log.Fatalln(err)
		}
	}
}
func (s *Shipwright[T]) parallelSteps(steps ...pipeline.Step[pipeline.Action]) error {
	if err := s.validateSteps(steps...); err != nil {
		return err
	}

	list := s.newList(steps...)
	list.Dependencies = s.prev

	if err := s.Collection.AddSteps(s.pipeline, list); err != nil {
		return fmt.Errorf("error adding step '%d' to collection. error: %w", list.Serial, err)
	}

	s.prev = []pipeline.Step[pipeline.StepList]{list}

	return nil
}
func (s *Shipwright[T]) parallelPipelines(steps ...pipeline.Step[pipeline.Pipeline]) error {
	return nil
}

func (s *Shipwright[T]) Cache(action pipeline.Action, c pipeline.Cacher) pipeline.Action {
	return action
}

func (s *Shipwright[T]) setup(steps ...pipeline.Step[T]) []pipeline.Step[T] {
	// if len(s.prev) > 0 {
	// 	for i := range steps {
	// 		if steps[i].Type != pipeline.StepTypeBackground {
	// 			steps[i].Dependencies = s.prev
	// 		}
	// 	}
	// }

	for i, step := range steps {
		// Set a default image for steps that don't provide one.
		// Most pre-made steps like `yarn`, `node`, `go` steps should provide a separate default image with those utilities installed.
		if step.Image == "" {
			image := plumbing.DefaultImage(s.Version)
			steps[i] = step.WithImage(image)
		}

		// Set a serial / unique identifier for this step so that we can reference it using the '-step' argument consistently.
		steps[i].Serial = s.n
		s.n++
	}

	return steps
}

func formatError(step pipeline.Step[pipeline.Action], err error) error {
	name := step.Name
	if name == "" {
		name = fmt.Sprintf("unnamed-step-%d", step.Serial)
	}

	return fmt.Errorf("[name: %s, id: %d] %w", name, step.Serial, err)
}

func (s *Shipwright[T]) validateSteps(steps ...pipeline.Step[pipeline.Action]) error {
	for _, v := range steps {
		err := s.Client.Validate(v)
		if err == nil {
			continue
		}

		if errors.Is(err, plumbing.ErrorSkipValidation) {
			s.Log.Warnln(formatError(v, err).Error())
			continue
		}

		return formatError(v, err)
	}

	return nil
}

func (s *Shipwright[T]) watchSignals() error {
	sig := cmdutil.WatchSignals()

	return fmt.Errorf("received OS signal: %s", sig.String())
}

// Execute is the equivalent of Done, but returns an error.
// Done should be preferred in Shipwright pipelines as it includes sub-process handling and logging.
func (s *Shipwright[T]) Execute(ctx context.Context) error {
	var (
		collection = s.Collection
	)
	// If the user has specified a specific step, then cut the "Collection" to only include that step
	if s.Opts.Args.Step != nil {
		step, err := collection.BySerial(*s.Opts.Args.Step)
		if err != nil {
			return fmt.Errorf("could not find step. Error: %w", err)
		}

		collection = collection.Sub(step)
	}

	if err := s.Client.Done(ctx, collection, s.Events); err != nil {
		return err
	}

	return nil
}

func (s *Shipwright[T]) Done() {
	var (
		ctx = context.Background()
	)

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.Opts.Tracer, "shipwright build")
	defer span.Finish()

	logger := s.Log.WithFields(plog.Combine(plog.TracingFields(ctx), plog.PipelineFields(s.Opts)))

	go func(logger *logrus.Entry) {
		if err := s.watchSignals(); err != nil {
			logger.WithFields(logrus.Fields{
				"status":       "cancelled",
				"completed_at": time.Now().Unix(),
			}).WithError(err).Errorln("execution completed")

			span.Finish()

			os.Exit(1)
		}
	}(logger)

	logger.WithField("mode", s.Opts.Args.Mode).Info("execution started")

	if err := s.Execute(ctx); err != nil {
		logger.WithFields(logrus.Fields{
			"status":       "error",
			"completed_at": time.Now().Unix(),
		}).WithError(err).Error("execution completed")
		return
	}

	logger.WithFields(logrus.Fields{
		"status":       "success",
		"completed_at": time.Now().Unix(),
	}).Info("execution completed")

	if v, ok := s.Opts.Tracer.(*jaeger.Tracer); ok {
		v.Close()
	}
}

type MultiFunc func(Shipwright[pipeline.Action])

// New creates a new Shipwright client which is used to create pipeline a single pipeline with many steps.
// This function will panic if the arguments in os.Args do not match what's expected.
// This function, and the type it returns, are only ran inside of a Shipwright pipeline, and so it is okay to treat this like it is the entrypoint of a command.
// Watching for signals, parsing command line arguments, and panics are all things that are OK in this function.
// New is used when creating a single pipeline. In order to create multiple pipelines, use the NewMulti function.
func New(name string) Shipwright[pipeline.Action] {
	args, err := plumbing.ParseArguments(os.Args[1:])
	if err != nil {
		log.Fatalln("Error parsing arguments. Error:", err)
	}

	if args == nil {
		log.Fatalln("Arguments list must not be nil")
		return Shipwright[pipeline.Action]{}
	}

	// Create standard packages based on the arguments provided.
	// This would be a good place to initialize loggers, tracers, etc
	var tracer opentracing.Tracer = &opentracing.NoopTracer{}

	logger := plog.New(args.LogLevel)
	jaegerCfg, err := config.FromEnv()
	if err == nil {
		// Here we ignore the closer because the jaegerTracer is the closer and we will just close that.
		jaegerTracer, _, err := jaegerCfg.NewTracer(config.Logger(jaeger.StdLogger))
		if err == nil {
			logger.Infoln("Initialized jaeger tracer")
			tracer = jaegerTracer
		} else {
			logger.Infoln("Could not initialize jaeger tracer; using no-op tracer; Error:", err.Error())
		}
	}

	sw := NewFromOpts(pipeline.CommonOpts{
		Name:    name,
		Version: args.Version,
		Output:  os.Stdout,
		Args:    args,
		Log:     logger,
		Tracer:  tracer,
	})

	// Ensure that no matter the behavior of the initializer, we still set the version on the shipwright object.
	sw.Version = args.Version
	sw.pipeline = DefaultPipelineID

	return sw
}

func NewFromOpts(opts pipeline.CommonOpts) Shipwright[pipeline.Action] {
	return NewClient[pipeline.Action](opts)
}

// NewWithClient creates a new Shipwright object with a specific client implementation.
// This function is intended to be used in very specific environments, like in tests.
func NewWithClient[T pipeline.StepContent](opts pipeline.CommonOpts, client pipeline.Client) Shipwright[T] {
	if opts.Args == nil {
		opts.Args = &plumbing.PipelineArgs{}
	}

	return Shipwright[T]{
		Client:     client,
		Opts:       opts,
		Log:        opts.Log,
		Collection: NewDefaultCollection(opts),
		pipeline:   DefaultPipelineID,

		n: 1,
	}
}
