package scribe

import (
	"context"

	"github.com/grafana/scribe/v2/dag"
)

func NewDaggerClient(ctx context.Context) (Client, error) {
	return &DaggerClient{}, nil
}

type DaggerClient struct{}

func (d *DaggerClient) Run(g *dag.Graph[Pipeline], opts ClientRunOpts) error {
	return nil
}
