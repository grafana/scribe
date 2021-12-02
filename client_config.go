package shipwright

import "pkg.grafana.com/shipwright/v1/types"

type ConfigClient struct{}

// Run will run the listed steps sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *ConfigClient) Run(_ ...types.Step) {
	panic("not implemented") // TODO: Implement
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *ConfigClient) Parallel(_ ...types.Step) {
	panic("not implemented") // TODO: Implement
}

func (c *ConfigClient) Cache(_ types.Step, _ types.Cacher) types.Step {
	panic("not implemented") // TODO: Implement
}

func (c *ConfigClient) Input(_ ...Argument) {
	panic("not implemented") // TODO: Implement
}

func (c *ConfigClient) Output(_ ...Output) {
	panic("not implemented") // TODO: Implement
}
