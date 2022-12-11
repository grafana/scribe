package drone

import (
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
)

func New(opts clients.CommonOpts) (pipeline.Client, error) {
	return &Client{
		Opts: opts,
		Log:  opts.Log,
	}, nil
}
