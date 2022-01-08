package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type DroneClient struct {
	Opts  *CommonOpts
	Root  *types.StepNode
	TreeN *types.StepNode
}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *DroneClient) Run(steps ...types.Step) {
	for i := range steps {
		node := &types.StepNode{
			Step: steps[i],
		}
		if c.Root == nil {
			c.Root = node
			c.TreeN = node
			continue
		}

		if err := c.TreeN.AppendChild(node); err != nil {
			plog.Infoln("Error appending child to graph", err)
		}

		c.TreeN = node
	}
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *DroneClient) Parallel(_ ...types.Step)                                  {}
func (c *DroneClient) Cache(_ types.StepAction, _ types.Cacher) types.StepAction { return nil }
func (c *DroneClient) Input(_ ...Argument)                                       {}
func (c *DroneClient) Output(_ ...Output)                                        {}
func (c *DroneClient) Init(_ *CommonOpts)                                        {}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (c *DroneClient) Done() {}

func NewDroneClient(opts *CommonOpts) Shipwright {
	return Shipwright{
		Client: &DroneClient{
			Opts: opts,
		},
	}
}
