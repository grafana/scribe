package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

type BuildOptions struct {
	// Names is the list of names / tags (including domain name) of the being created. Of course, it could be anything and doesn't have to include the domain name.
	// Names should include the tag (the string that follows the colon).
	// Examples:
	// "grafana/shipwright:latest"
	// "docker.io/grafana/shipwright:v1.0.0"
	Names []string

	// ContextDir is packaged into a tarball and provided to the docker daemon as the docker build context. Read more about docker build contexts here:
	// https://docs.docker.com/engine/reference/commandline/build/
	// If a '.dockerignore' is located in the root of the provided ContextDir, then it is parsed and used.
	ContextDir string

	// Dockerfile is the path to the Dockerfile used to build this image. It does not have to be in the ContextDir.
	Dockerfile string

	// Args are BuildArgs provided to the docker build process. Read more about docker build args here:
	// https://docs.docker.com/engine/reference/commandline/build/#set-build-time-variables---build-arg
	Args map[string]*string

	Stdout io.Writer
}

func BuildStep(buildOpts BuildOptions) pipeline.Step {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		buildOpts.Stdout = opts.Stdout
		return Build(ctx, buildOpts)
	})
}

func Build(ctx context.Context, opts BuildOptions) error {
	client := dockerClient()

	if opts.ContextDir == "" {
		opts.ContextDir = "."
	}

	buildContext, err := ImageContext(opts.ContextDir)
	if err != nil {
		return err
	}

	res, err := client.ImageBuild(ctx, buildContext, types.ImageBuildOptions{
		Tags:       opts.Names,
		BuildArgs:  opts.Args,
		Dockerfile: opts.Dockerfile,
	})

	if err != nil {
		return err
	}
	defer res.Body.Close()

	return WriteBuildLogs(res.Body, opts.Stdout)
}
