package shipwright

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/plog"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// Shipwright is the client that is used in every pipeline to declare the steps that make up a pipeline.
type Shipwright struct {
	Client     pipeline.Client
	Collection pipeline.Collection

	pipeline.Configurer

	// Opts are the options that are provided to the pipeline from outside sources. This includes mostly command-line arguments and environment variables
	Opts pipeline.CommonOpts
	Log  *logrus.Logger

	// n tracks the ID of a step so that the "shipwright -step=" argument will function independently of the client implementation
	// It ensures that the 11th step in a Drone generated pipeline is also the 11th step in a CLI pipeline
	n       int
	Version string
}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (s *Shipwright) Run(step ...pipeline.Step) {
	steps := s.Setup(step...)

	if err := s.validateSteps(steps...); err != nil {
		s.Log.Fatalln(err)
	}

	for i := range steps {
		if err := s.Collection.Append(steps[i]); err != nil {
			s.Log.Fatalln(err)
		}
	}
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (s *Shipwright) Parallel(step ...pipeline.Step) {
	steps := s.Setup(step...)

	if err := s.validateSteps(steps...); err != nil {
		s.Log.Fatalln(err)
	}

	if err := s.Collection.Append(steps...); err != nil {
		s.Log.Fatalln(err)
	}
}

// These functions are just ideas at the moment.
// // Go is the equivalent of `go func()`. This function will run a step asynchronously and continue on to the next.
// // Go(...pipeline.Step)
// // func (s *Shipwright) Input(...pipeline.Argument) {}
// // func (s *Shipwright) Output(...pipeline.Output) {}

func (s *Shipwright) Cache(action pipeline.StepAction, c pipeline.Cacher) pipeline.StepAction {
	return action
}

func (s *Shipwright) Setup(steps ...pipeline.Step) []pipeline.Step {
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

func formatError(step pipeline.Step, err error) error {
	name := step.Name
	if name == "" {
		name = fmt.Sprintf("unnamed-step-%d", step.Serial)
	}

	return fmt.Errorf("[name: %s, id: %d] %w", name, step.Serial, err)
}

func (s *Shipwright) validateSteps(steps ...pipeline.Step) error {
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

func (s *Shipwright) Done() {
	var (
		ctx        = context.Background()
		collection = s.Collection
	)

	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.Opts.Tracer, "shipwright build")
	defer span.Finish()

	logger := s.Log.WithFields(plog.Combine(plog.TracingFields(ctx), plog.PipelineFields(s.Opts)))

	logger.WithField("mode", s.Opts.Args.Mode).Info("execution started")

	// If the user has specified a specific step, then cut the "Collection" to only include that step
	if s.Opts.Args.Step != nil {
		step, err := collection.BySerial(*s.Opts.Args.Step)
		if err != nil {
			logger.Panicln("could not find step", err)
		}

		logger.Infoln("Found step at", *s.Opts.Args.Step, "named", step.Name)

		collection = collection.Sub(step)
	}

	if err := s.Client.Done(ctx, collection); err != nil {
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

// New creates a new Shipwright client which is used to create pipeline steps.
// This function will panic if the arguments in os.Args do not match what's expected.
// This function, and the type it returns, are only ran inside of a Shipwright pipeline, and so it is okay to treat this like it is the entrypoint of a command.
// Watching for signals, parsing command line arguments, and panics are all things that are OK in this function.
func New(name string, events ...pipeline.Event) Shipwright {
	args, err := plumbing.ParseArguments(os.Args[1:])
	if err != nil {
		log.Fatalln("Error parsing arguments. Error:", err)
	}

	if args == nil {
		log.Fatalln("Arguments list must not be nil")
		return Shipwright{}
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

	return sw
}

func NewFromOpts(opts pipeline.CommonOpts, events ...pipeline.Event) Shipwright {
	return NewClient(opts)
}
