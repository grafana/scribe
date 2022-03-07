package pipeline

import (
	"io"

	"github.com/sirupsen/logrus"
	"github.com/grafana/shipwright/plumbing"
)

// CommonOpts are provided in the Client's Init function, which includes options that are common to all clients, like
// logging, output, and debug options
type CommonOpts struct {
	Name    string
	Version string
	Output  io.Writer
	Args    *plumbing.PipelineArgs
	Log     *logrus.Logger
}
