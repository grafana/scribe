package shipwright

import "pkg.grafana.com/shipwright/v1/types"

// The CLIClient is used when interacting with a shipwright pipeline using the shipwright CLI
type CLIClient struct{}

func (c *CLIClient) Cache(_ types.Step, _ types.Cacher) types.Step {
	return nil
}

func (c *CLIClient) Input(_ ...Argument) {}
func (c *CLIClient) Output(_ ...Output)  {}

// Parallel executes the provided steps at the same time.
func (c *CLIClient) Parallel(steps ...types.Step) {}
func (c *CLIClient) Run(steps ...types.Step)      {}
