package graphviz

import (
	"context"
	"io"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
)

type Client struct {
	Stdout io.Writer
}

func New(ctx context.Context, opts clients.CommonOpts) (pipeline.Client, error) {
	return &Client{
		Stdout: opts.Output,
	}, nil
}

func (c *Client) Done(ctx context.Context, w *pipeline.Collection) error {
	pipelines := []pipeline.Pipeline{}
	if err := w.WalkPipelines(ctx, func(ctx context.Context, p pipeline.Pipeline) error {
		pipelines = append(pipelines, p)
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (c *Client) Validate(step pipeline.Step) error {
	return nil
}
