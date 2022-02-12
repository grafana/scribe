package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/docker"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/drone"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

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
		Client: &drone.Client{
			Log:  opts.Log,
			List: pipeline.NewList(),
			Opts: opts,
		},
	}
}

func NewCLIClient(opts pipeline.CommonOpts) Shipwright {
	return Shipwright{
		Client: &cli.Client{
			Log:   opts.Log,
			Opts:  opts,
			Queue: &pipeline.StepQueue{},
		},
	}
}

func NewDockerClient(opts pipeline.CommonOpts) Shipwright {
	return Shipwright{
		Client: &docker.Client{
			Log:   opts.Log,
			Opts:  opts,
			Queue: &pipeline.StepQueue{},
		},
	}
}
