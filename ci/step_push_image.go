package main

import (
	"context"
	"fmt"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
)

var (
	ArgumentDockerUsername = pipeline.NewSecretArgument("docker_username")
	ArgumentDockerPassword = pipeline.NewSecretArgument("docker_password")
)

func StepPushImage(version string, image Image) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		var (
			client       = Client()
			imageWithTag = image.Tag(opts.State.MustGetString(ArgumentVersion))
			username     = opts.State.MustGetString(ArgumentDockerUsername)
			password     = opts.State.MustGetString(ArgumentDockerPassword)

			s    = strings.Split(imageWithTag, ":")
			name = s[0]
			tag  = s[1]
		)

		opts.Logger.Infoln("Pushing", name, "with tag", tag)

		return client.PushImage(docker.PushImageOptions{
			Name:          name,
			Tag:           tag,
			Registry:      plumbing.DefaultRegistry(),
			RawJSONStream: false,
			OutputStream:  opts.Stdout,
		}, docker.AuthConfiguration{
			Username: username,
			Password: password,
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS, ArgumentDockerUsername, ArgumentDockerPassword, ArgumentVersion).
		WithImage(plumbing.SubImage("docker", version))
}

func PushSteps(version string, images []Image) []pipeline.Step {
	steps := make([]pipeline.Step, len(images))

	for i, image := range images {
		steps[i] = StepPushImage(version, image).WithName(fmt.Sprintf("push %s", image.Name))
	}

	return steps
}
