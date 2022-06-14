package main

import (
	"context"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func ListImages() pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		client := Client()
		images, err := client.ListImages(docker.ListImagesOptions{
			Context: ctx,
			Filters: map[string][]string{
				"source": {"scribe"},
			},
		})
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
