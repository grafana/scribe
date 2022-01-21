package types

import (
	"io"

	"pkg.grafana.com/shipwright/v1/plumbing"
)

// CommonOpts are provided in the Client's Init function, which includes options that are common to all clients, like
// logging, output, and debug options
type CommonOpts struct {
	Name   string
	Output io.Writer
	Args   *plumbing.Arguments
}
