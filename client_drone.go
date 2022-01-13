package shipwright

import (
	"gopkg.in/yaml.v2"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/schemas/drone"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type DroneClient struct {
	Opts *CommonOpts
	List *types.List
}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *DroneClient) Run(steps ...types.Step) {

	c.List.AppendLineage(steps...)
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *DroneClient) Parallel(steps ...types.Step) {
	c.List.Append(steps...)
}

func (c *DroneClient) Cache(_ types.StepAction, _ types.Cacher) types.StepAction { return nil }
func (c *DroneClient) Input(_ ...Argument)                                       {}
func (c *DroneClient) Output(_ ...Output)                                        {}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *DroneClient) Done() {
	cfg := &drone.Pipeline{
		Name: c.Opts.Name,
		Kind: "pipeline",
		Type: "docker",
		Clone: drone.CloneSettings{
			Disable: true,
		},
		Steps: []drone.Step{},
	}

	c.List.Walk(func(s types.Step) error {
		cfg.Steps = append(cfg.Steps, drone.NewStep(c.Opts.Args.Path, s))
		return nil
	})

	if err := yaml.NewEncoder(c.Opts.Output).Encode(cfg); err != nil {
		plog.Fatalln(err)
	}
}
func NewDroneClient(opts *CommonOpts) Shipwright {
	return Shipwright{
		Client: &DroneClient{
			List: types.NewList(),
			Opts: opts,
		},
	}
}
