package scribe

import (
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/scribe/plumbing/pipeline/clients/cli"
	"github.com/grafana/scribe/plumbing/pipeline/clients/dagger"
	"github.com/grafana/scribe/plumbing/pipeline/clients/drone"
)

var (
	// ClientCLI is set when a pipeline is ran from the Scribe CLI, typically for local development, but can also be set when running Scribe within a third-party service like CircleCI or Drone
	ClientCLI string = "cli"

	// ClientDrone is set when a pipeline is ran using the Drone client, which is used to generate a Drone config from a Scribe pipeline
	ClientDrone = "drone"

	// ClientDagger
	ClientDagger = "dagger"
)

func NewDefaultCollection(opts pipeline.CommonOpts) *pipeline.Collection {
	p := pipeline.NewCollection()
	if err := p.AddPipelines(pipeline.New(opts.Name, DefaultPipelineID)); err != nil {
		panic(err)
	}

	return p
}

func NewMultiCollection() *pipeline.Collection {
	return pipeline.NewCollection()
}

type InitializerFunc func(pipeline.CommonOpts) pipeline.Client

// The ClientInitializers define how different RunModes initialize the Scribe client
var ClientInitializers = map[string]InitializerFunc{
	ClientCLI:    NewCLIClient,
	ClientDrone:  NewDroneClient,
	ClientDagger: NewDaggerClient,
}

func NewDroneClient(opts pipeline.CommonOpts) pipeline.Client {
	return &drone.Client{
		Opts: opts,
		Log:  opts.Log,
	}
}

func NewCLIClient(opts pipeline.CommonOpts) pipeline.Client {
	return &cli.Client{
		Opts: opts,
		Log:  opts.Log,
	}
}

func NewDaggerClient(opts pipeline.CommonOpts) pipeline.Client {
	return &dagger.Client{
		Opts: opts,
		Log:  opts.Log,
	}
}

func RegisterClient(name string, initializer InitializerFunc) {
	ClientInitializers[name] = initializer
}
