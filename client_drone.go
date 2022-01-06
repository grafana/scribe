package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type ClientDrone struct{}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *ClientDrone) Run(_ ...types.Step) {
	panic("not implemented") // TODO: Implement
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *ClientDrone) Parallel(_ ...types.Step) {
	panic("not implemented") // TODO: Implement
}

func (c *ClientDrone) Cache(_ types.Step, _ types.Cacher) types.Step {
	panic("not implemented") // TODO: Implement
}

func (c *ClientDrone) Input(_ ...Argument) {
	panic("not implemented") // TODO: Implement
}

func (c *ClientDrone) Output(_ ...Output) {
	panic("not implemented") // TODO: Implement
}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (c *ClientDrone) Done() {
	panic("not implemented") // TODO: Implement
}

// Parse parses the CLI flags provided to the pipeline
func (c *ClientDrone) Parse(args []string) error {
	panic("not implemented") // TODO: Implement
}

func NewDroneClient() Shipwright {
	return Shipwright{
		Client: &ClientDrone{},
	}
}
