package scribe

import (
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/pipeline/clients/cli"
	"github.com/grafana/scribe/pipeline/clients/dagger"
	"github.com/grafana/scribe/pipeline/clients/drone"
)

var (
	// ClientCLI is set when a pipeline is ran from the Scribe CLI, typically for local development, but can also be set when running Scribe within a third-party service like CircleCI or Drone
	ClientCLI string = "cli"

	// ClientDrone is set when a pipeline is ran using the Drone client, which is used to generate a Drone config from a Scribe pipeline
	ClientDrone = "drone"

	// ClientDagger
	ClientDagger = "dagger"
)

func NewDefaultCollection(opts clients.CommonOpts) *pipeline.Collection {
	p := pipeline.NewCollection()
	if err := p.AddPipelines(pipeline.New(opts.Name, DefaultPipelineID)); err != nil {
		panic(err)
	}
	p.Root = []int64{DefaultPipelineID}

	return p
}

func NewMultiCollection() *pipeline.Collection {
	return pipeline.NewCollection()
}

type InitializerFunc func(clients.CommonOpts) (pipeline.Client, error)

// The ClientInitializers define how different RunModes initialize the Scribe client
var ClientInitializers = map[string]InitializerFunc{
	ClientCLI:    cli.New,
	ClientDrone:  drone.New,
	ClientDagger: dagger.New,
}

func RegisterClient(name string, initializer InitializerFunc) {
	ClientInitializers[name] = initializer
}
