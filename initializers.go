package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/drone"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

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
