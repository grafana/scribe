package scribe

import (
	"context"

	"github.com/grafana/scribe/v2/dag"
)

// ClientInitializerFunc is a function that initializes a client.
type ClientInitializerFunc func(ctx context.Context) (Client, error)

type ClientRunOpts struct{}

// Client handles the given dag.Graph in the appropriate way for that client.
// For example, a client for something like Drone or CircleCI might look at the contents of the pipeline and generate a file to the given writer or directory.
// But the Drone client will actually execute the pipeline.
type Client interface {
	Run(*dag.Graph[Pipeline], ClientRunOpts) error
}

var Clients = map[string]ClientInitializerFunc{
	"dagger": NewDaggerClient,
}
