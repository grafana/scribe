package shipwright

import (
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/docker"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/drone"
)

func NewDefaultCollection(opts pipeline.CommonOpts) *pipeline.Collection {
	p := pipeline.NewCollection()

	p.AddPipelines(pipeline.Step[pipeline.Pipeline]{
		Name:   opts.Name,
		Serial: DefaultPipelineID,
	})

	return p
}

func NewMultiCollection() *pipeline.Collection {
	return pipeline.NewCollection()
}

type InitializerFunc func(pipeline.CommonOpts) pipeline.Client

// The ClientInitializers define how different RunModes initialize the Shipwright client
var ClientInitializers = map[plumbing.RunModeOption]InitializerFunc{
	plumbing.RunModeCLI:    NewCLIClient,
	plumbing.RunModeDrone:  NewDroneClient,
	plumbing.RunModeConfig: NewCLIClient,
	plumbing.RunModeServer: NewCLIClient,
	plumbing.RunModeDocker: NewDockerClient,
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

func NewDockerClient(opts pipeline.CommonOpts) pipeline.Client {
	return &docker.Client{
		Opts: opts,
		Log:  opts.Log,
	}
}
