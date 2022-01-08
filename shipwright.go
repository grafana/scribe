package shipwright

import (
	"os"

	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/golang"
	"pkg.grafana.com/shipwright/v1/make"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
	"pkg.grafana.com/shipwright/v1/yarn"
)

type Argument interface{}
type Output interface{}

type Artifact struct {
	Ref      string
	Artifact types.Artifact
}

func NewArtifact(ref string, artifact types.Artifact) Artifact {
	return Artifact{}
}

type Client interface {
	// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
	// This function blocks the goroutine until all of the steps have completed.
	Run(...types.Step)

	// Parallel will run the listed steps at the same time.
	// This function blocks the goroutine until all of the steps have completed.
	Parallel(...types.Step)

	Cache(types.StepAction, types.Cacher) types.StepAction

	Input(...Argument)
	Output(...Output)

	// Done must be ran at the end of the pipeline.
	// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
	Done()

	// Init initalizes the client with the common options
	Init(*CommonOpts)
}

type Shipwright struct {
	Client
	Git    git.Client
	FS     fs.Client
	Golang golang.Client
	Make   make.Client
	Yarn   yarn.Client
}

// New creates a new Shipwright client which is used to create pipeline steps.
// This function will panic if the arguments in os.Args do not match what's expected.
func New(name string, events ...types.Event) Shipwright {
	args, err := plumbing.ParseArguments(os.Args)
	if err != nil {
		plog.Fatalln("Error parsing arguments. Error:", err)
	}

	if args == nil {
		plog.Fatalln("Arguments list must not be nil")
	}

	return NewFromOpts(&CommonOpts{
		Name:   name,
		Output: os.Stdout,
		Args:   args,
	})
}

func NewFromOpts(opts *CommonOpts, events ...types.Event) Shipwright {
	return opts.NewClient()
}
