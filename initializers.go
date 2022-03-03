package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline/clients/docker"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline/clients/drone"
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
