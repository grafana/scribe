package shipwright

import (
	"pkg.grafana.com/shipwright/v1/docker"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/golang"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

// NewClient creates a new Shipwright client based on the commonopts (mostly the mode).
// It does not check for a non-nil "Args" field.
func NewClient(c *pipeline.CommonOpts) Shipwright {
	c.Log.Infof("Initializing Shipwright client with mode '%s'", c.Args.Mode.String())
	initializer, ok := ClientInitializers[c.Args.Mode]
	if !ok {
		c.Log.Fatalln("Could not initialize shipwright. Could not find initializer for mode", c.Args.Mode)
		return Shipwright{}
	}

	s := initializer(c)

	// Initialize the individual clients now
	s.Git = git.New(s, c)
	s.Docker = docker.New(c)
	s.Golang = golang.New(c)
	return s
}
