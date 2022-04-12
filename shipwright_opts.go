package shipwright

import (
	"github.com/grafana/shipwright/plumbing/pipeline"
)

// NewClient creates a new Shipwright client based on the commonopts (mostly the mode).
// It does not check for a non-nil "Args" field.
func NewClient[T pipeline.StepContent](c pipeline.CommonOpts) Shipwright[T] {
	c.Log.Infof("Initializing Shipwright client with mode '%s'", c.Args.Mode.String())
	sw := Shipwright[T]{}

	initializer, ok := ClientInitializers[c.Args.Mode]
	if !ok {
		c.Log.Fatalln("Could not initialize shipwright. Could not find initializer for mode", c.Args.Mode)
		return Shipwright[T]{}
	}

	sw.Client = initializer(c)

	// TODO: initialize the collection based on other factors and not just use the default one.
	sw.Collection = NewDefaultCollection(c)

	sw.Opts = c
	sw.Log = c.Log

	return sw
}
