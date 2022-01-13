package shipwright

import (
	"io"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

var ClientInitializers = map[plumbing.RunModeOption]func(*CommonOpts) Shipwright{
	plumbing.RunModeCLI:    NewCLIClient,
	plumbing.RunModeDrone:  NewDroneClient,
	plumbing.RunModeConfig: NewCLIClient,
	plumbing.RunModeServer: NewCLIClient,
	plumbing.RunModeDocker: NewCLIClient,
}

// CommonOpts are provided in the Client's Init function, which includes options that are common to all clients, like
// logging, output, and debug options
type CommonOpts struct {
	Name   string
	Output io.Writer
	Args   *plumbing.Arguments
}

// NewClient creates a new Shipwright client based on the commonopts (mostly the mode).
// It does not check for a non-nil "Args" field.
func (c *CommonOpts) NewClient() Shipwright {
	plog.Infof("Initializing Shipwright client with mode '%s'", c.Args.Mode.String())
	initializer, ok := ClientInitializers[c.Args.Mode]
	if !ok {
		plog.Fatalln("Could not initialize shipwright. Could not find initializer for mode", c.Args.Mode)
	}

	return initializer(c)
}
