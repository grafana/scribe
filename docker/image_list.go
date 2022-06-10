package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

func ListImages(ctx context.Context) ([]types.ImageSummary, error) {
	client := dockerClient()

	return client.ImageList(ctx, types.ImageListOptions{
		All: true,
	})
}
