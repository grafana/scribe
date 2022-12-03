package cli

import (
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
)

func New(opts clients.CommonOpts) pipeline.Client {
	return &Client{
		Opts: opts,
		Log:  opts.Log,
	}
}
