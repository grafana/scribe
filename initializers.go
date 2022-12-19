package scribe

import (
	"context"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/pipeline/clients/cli"
	"github.com/grafana/scribe/pipeline/clients/dagger"
	"github.com/grafana/scribe/pipeline/clients/drone"
	"github.com/grafana/scribe/pipeline/clients/graphviz"
)

var (
	ClientCLI      string = "cli"
	ClientDrone           = "drone"
	ClientDagger          = "dagger"
	ClientGraphviz        = "graphviz"
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

type InitializerFunc func(context.Context, clients.CommonOpts) (pipeline.Client, error)

// The ClientInitializers define how different RunModes initialize the Scribe client
var ClientInitializers = map[string]InitializerFunc{
	ClientCLI:      cli.New,
	ClientDrone:    drone.New,
	ClientDagger:   dagger.New,
	ClientGraphviz: graphviz.New,
}

func RegisterClient(name string, initializer InitializerFunc) {
	ClientInitializers[name] = initializer
}
