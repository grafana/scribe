package shipwright

import (
	"os"

	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/golang"
	"pkg.grafana.com/shipwright/v1/make"
	"pkg.grafana.com/shipwright/v1/types"
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
	// Run will run the listed steps sequentially. For example, the second step will not run until the first step has completed.
	// This function blocks the goroutine until all of the steps have completed.
	Run(...types.Step)

	// Parallel will run the listed steps at the same time.
	// This function blocks the goroutine until all of the steps have completed.
	Parallel(...types.Step)

	Cache(types.Step, types.Cacher) types.Step
	Input(...Argument)
	Output(...Output)
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
func New(events ...types.Event) Shipwright {
	opts, err := ParseCLIOpts(os.Args)
	if err != nil {
		panic(err)
	}

	return NewFromOpts(opts, events)
}

func NewFromOpts(opts *Opts, events ...types.Event) Shipwright {
	var client Client
	switch opts.Mode {
	case RunModeServer:
	case RunModeConfig:
		client = &ConfigClient{}
	default:
		client = &CLIClient{}
	}

	return Shipwright{
		Client: client,
	}
}
