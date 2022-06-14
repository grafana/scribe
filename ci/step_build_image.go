package main

import (
	"context"
	"fmt"

	docker "github.com/fsouza/go-dockerclient"
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
		client := Client()
		version := opts.State.MustGetString(ArgumentVersion)
		tag := image.Tag(version)

		opts.Logger.Infoln("Building", image.Dockerfile, "with tag", tag)

		return client.BuildImage(docker.BuildImageOptions{
			Context:    ctx,
			Name:       tag,
			Dockerfile: image.Dockerfile,
			ContextDir: image.Context,
			BuildArgs: []docker.BuildArg{
				{
					Name:  "VERSION",
					Value: version,
				},
			},
			Labels: map[string]string{
				"source": "scribe",
			},
			OutputStream: opts.Stdout,
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS, ArgumentVersion).
		WithImage(plumbing.SubImage("docker", version))
}
