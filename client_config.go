package shipwright

import (
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type ConfigClient struct{}

// Run will run the listed steps sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *ConfigClient) Run(_ ...types.Step) {}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *ConfigClient) Parallel(_ ...types.Step)                      {}
func (c *ConfigClient) Cache(_ types.Step, _ types.Cacher) types.Step { return nil }
func (c *ConfigClient) Input(_ ...Argument)                           {}
func (c *ConfigClient) Output(_ ...Output)                            {}
func (c *ConfigClient) Done()                                         {}
func (c *ConfigClient) Parse(args []string) error                     { return nil }
func NewConfigClient() Shipwright {
	return Shipwright{
		Client: &ConfigClient{},
	}
}
