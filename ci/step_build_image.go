package main

import (
	"context"
	"fmt"

	"github.com/grafana/scribe/docker"
	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func BuildSteps(version string, images []Image) []pipeline.Step {
	steps := make([]pipeline.Step, len(images))

	for i, image := range images {
		steps[i] = StepBuildImage(version, image).WithName(fmt.Sprintf("build %s image", image.Name))
	}

	return steps
}

func StepBuildImage(version string, image Image) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		version := opts.State.MustGetString(ArgumentVersion)
		tag := image.Tag(version)

		opts.Logger.Infoln("Building", image.Dockerfile, "with tag", tag)
		return docker.Build(ctx, docker.BuildOptions{
			Names:      []string{tag},
			Dockerfile: image.Dockerfile,
			ContextDir: image.Context,
			Args: map[string]*string{
				"VERSION": &version,
			},
			Stdout: opts.Stdout,
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS, ArgumentVersion).
		WithImage(plumbing.SubImage("docker", version))
}
