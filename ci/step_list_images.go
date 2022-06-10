package main

import (
	"context"

	"github.com/grafana/scribe/docker"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func ListImages() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		images, err := docker.ListImages(ctx)
		if err != nil {
			return err
		}

		for _, v := range images {
			opts.Logger.Infof("Got image: %10s | %32v | %10d", v.ID, v.RepoTags, v.Size)
		}

		return nil
	}

	return pipeline.NewStep(action).WithArguments(pipeline.ArgumentDockerSocketFS)
}
