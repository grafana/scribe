package main

import (
	"context"
	"fmt"

	"github.com/grafana/scribe/docker"
	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

func StepPushImage(version string, image Image) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		tag := image.Tag(opts.State.MustGetString(ArgumentVersion))

		auth, err := opts.State.GetString(ArgumentDockerAuthToken)
		if err != nil {
			return err
		}

		opts.Logger.Infoln("Pushing", tag)
		return docker.Push(ctx, docker.PushOpts{
			Name:      tag,
			Registry:  plumbing.DefaultRegistry(),
			AuthToken: auth,
			InfoOut:   opts.Stdout,
			DebugOut:  opts.Logger.WithField("action", "push").WriterLevel(logrus.DebugLevel),
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS, ArgumentDockerAuthToken, ArgumentVersion).
		WithImage(plumbing.SubImage("docker", version))
}

func PushSteps(version string, images []Image) []pipeline.Step {
	steps := make([]pipeline.Step, len(images))

	for i, image := range images {
		steps[i] = StepPushImage(version, image).WithName(fmt.Sprintf("push %s", image.Name))
	}

	return steps
}
