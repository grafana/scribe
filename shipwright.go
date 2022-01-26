package shipwright

import (
	"os"

	"pkg.grafana.com/shipwright/v1/docker"
	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/golang"
	"pkg.grafana.com/shipwright/v1/make"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/config"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
	"pkg.grafana.com/shipwright/v1/yarn"
)

type Client interface {
	config.Configurer

	// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
	// This function blocks the goroutine until all of the steps have completed.
	Run(...types.Step)

	// Parallel will run the listed steps at the same time.
	// This function blocks the goroutine until all of the steps have completed.
	Parallel(...types.Step)

	// Go is the equivalent of `go func()`. This function will run a step asynchronously and continue on to the next.
	// Go(...types.Step)

	Cache(types.StepAction, types.Cacher) types.StepAction

	Input(...types.Argument)
	Output(...types.Output)

	// Done must be ran at the end of the pipeline.
	// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
	Done()
}

type Shipwright struct {
	Client
	Git    git.Client
	FS     fs.Client
	Golang golang.Client
	Make   make.Client
	Yarn   yarn.Client
	Docker docker.Client

	// n tracks the ID of a step so that the "shipwright -step=" argument will function independently of the client implementation
	// It ensures that the 11th step in a Drone generated pipeline is also the 11th step in a CLI pipeline
	n int

	version string
}

func (s *Shipwright) initSteps(steps ...types.Step) []types.Step {
	for i, step := range steps {
		// Set a default image for steps that don't provide one.
		// Most pre-made steps like `yarn`, `node`, `go` steps should provide a separate default image with those utilities installed.
		if step.Image == "" {
			image := plumbing.DefaultImage(s.version)
			steps[i] = step.WithImage(image)
		}

		// Set a serial / unique identifier for this step so that we can reference it using the '-step' argument consistently.
		steps[i].Serial = s.n
		s.n++
	}

	return steps
}

func (s *Shipwright) Run(steps ...types.Step) {
	initializedSteps := s.initSteps(steps...)

	s.Client.Run(initializedSteps...)
}

func (s *Shipwright) Parallel(steps ...types.Step) {
	initializedSteps := s.initSteps(steps...)

	s.Client.Parallel(initializedSteps...)
}

// New creates a new Shipwright client which is used to create pipeline steps.
// This function will panic if the arguments in os.Args do not match what's expected.
func New(name string, events ...types.Event) Shipwright {
	args, err := plumbing.ParseArguments(os.Args[1:])
	if err != nil {
		plog.Fatalln("Error parsing arguments. Error:", err)
	}

	if args == nil {
		plog.Fatalln("Arguments list must not be nil")
		return Shipwright{}
	}

	sw := NewFromOpts(&types.CommonOpts{
		Name:    name,
		Version: args.Version,
		Output:  os.Stdout,
		Args:    args,
	})

	// Ensure that no matter the behavior of the initializer, we still set the version on the shipwright object.
	sw.version = args.Version

	return sw
}

func NewFromOpts(opts *types.CommonOpts, events ...types.Event) Shipwright {
	return NewClient(opts)
}
