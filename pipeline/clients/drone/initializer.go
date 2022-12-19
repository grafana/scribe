package drone

import (
	"context"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
)

func New(ctx context.Context, opts clients.CommonOpts) (pipeline.Client, error) {
	return &Client{
		Opts: opts,
		Log:  opts.Log,
	}, nil
}
