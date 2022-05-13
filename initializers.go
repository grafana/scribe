package shipwright

import (
	"github.com/docker/docker/client"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/docker"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/drone"
)

func NewDefaultCollection(opts pipeline.CommonOpts) *pipeline.Collection {
	p := pipeline.NewCollection()
	if err := p.AddPipelines(pipeline.NewPipelineNode(opts.Name, DefaultPipelineID)); err != nil {
		panic(err)
	}

	return p
}

func NewMultiCollection() *pipeline.Collection {
	return pipeline.NewCollection()
}

type InitializerFunc func(pipeline.CommonOpts) pipeline.Client

// The ClientInitializers define how different RunModes initialize the Shipwright client
var ClientInitializers = map[plumbing.RunModeOption]InitializerFunc{
	plumbing.RunModeCLI:           NewCLIClient,
	plumbing.RunModeDrone:         NewDroneClient,
	plumbing.RunModeConfig:        NewCLIClient,
	plumbing.RunModeServer:        NewCLIClient,
	plumbing.RunModeDocker:        NewDockerClient,
	plumbing.RunModeDroneStarlark: NewDroneStarlarkClient,
}

func NewDroneClient(opts pipeline.CommonOpts) pipeline.Client {
	return &drone.Client{
		Opts:     opts,
		Log:      opts.Log,
		Language: drone.LanguageYAML,
	}
}

func NewDroneStarlarkClient(opts pipeline.CommonOpts) pipeline.Client {
	return &drone.Client{
		Opts:     opts,
		Log:      opts.Log,
		Language: drone.LanguageStarlark,
	}
}

func NewCLIClient(opts pipeline.CommonOpts) pipeline.Client {
	return &cli.Client{
		Opts: opts,
		Log:  opts.Log,
	}
}

func NewDockerClient(opts pipeline.CommonOpts) pipeline.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &docker.Client{
		Client: cli,
		Opts:   opts,
		Log:    opts.Log,
	}
}
