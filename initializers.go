package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/docker"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/drone"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// The ClientInitializers define how different RunModes initialize the Shipwright client
var ClientInitializers = map[plumbing.RunModeOption]func(*types.CommonOpts) Shipwright{
	plumbing.RunModeCLI:    NewCLIClient,
	plumbing.RunModeDrone:  NewDroneClient,
	plumbing.RunModeConfig: NewCLIClient,
	plumbing.RunModeServer: NewCLIClient,
	plumbing.RunModeDocker: NewDockerClient,
}

func NewDroneClient(opts *types.CommonOpts) Shipwright {
	return Shipwright{
		Client: &drone.Client{
			List: types.NewList(),
			Opts: opts,
		},
	}
}

func NewCLIClient(opts *types.CommonOpts) Shipwright {
	return Shipwright{
		Client: &cli.Client{
			Opts:  opts,
			Queue: &types.StepQueue{},
		},
	}
}

func NewDockerClient(opts *types.CommonOpts) Shipwright {
	return Shipwright{
		Client: &docker.Client{
			Opts:  opts,
			Queue: &types.StepQueue{},
		},
	}
}
