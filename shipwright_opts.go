package shipwright

import (
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// The ClientInitializers define how different RunModes initialize the Shipwright client
var ClientInitializers = map[plumbing.RunModeOption]func(*types.CommonOpts) Shipwright{
	plumbing.RunModeCLI:    NewCLIClient,
	plumbing.RunModeDrone:  NewDroneClient,
	plumbing.RunModeConfig: NewCLIClient,
	plumbing.RunModeServer: NewCLIClient,
	plumbing.RunModeDocker: NewCLIClient,
}

// NewClient creates a new Shipwright client based on the commonopts (mostly the mode).
// It does not check for a non-nil "Args" field.
func NewClient(c *types.CommonOpts) Shipwright {
	plog.Infof("Initializing Shipwright client with mode '%s'", c.Args.Mode.String())
	initializer, ok := ClientInitializers[c.Args.Mode]
	if !ok {
		plog.Fatalln("Could not initialize shipwright. Could not find initializer for mode", c.Args.Mode)
	}

	s := initializer(c)

	// Initialize the individual clients now
	s.Git = git.New(s)

	return s
}
