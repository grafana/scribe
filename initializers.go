package shipwright

import (
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/docker"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/drone"
)

func NewDefaultCollection() pipeline.Collection {
	return pipeline.NewQueue()
}

// The ClientInitializers define how different RunModes initialize the Shipwright client
var ClientInitializers = map[plumbing.RunModeOption]func(pipeline.CommonOpts) Shipwright{
	plumbing.RunModeCLI:    NewCLIClient,
	plumbing.RunModeDrone:  NewDroneClient,
	plumbing.RunModeConfig: NewCLIClient,
	plumbing.RunModeServer: NewCLIClient,
	plumbing.RunModeDocker: NewDockerClient,
}

func NewDroneClient(opts pipeline.CommonOpts) Shipwright {
	return Shipwright{
		Collection: NewDefaultCollection(),
		Client: &drone.Client{
			Opts: opts,
			Log:  opts.Log,
		},
	}
}

func NewCLIClient(opts pipeline.CommonOpts) Shipwright {
	return Shipwright{
		Collection: NewDefaultCollection(),
		Client: &cli.Client{
			Opts: opts,
			Log:  opts.Log,
		},
	}
}

func NewDockerClient(opts pipeline.CommonOpts) Shipwright {
	return Shipwright{
		Collection: NewDefaultCollection(),
		Client: &docker.Client{
			Opts: opts,
			Log:  opts.Log,
		},
	}
}
