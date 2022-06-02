package docker

import (
	"context"
	"fmt"

	"github.com/grafana/scribe/docker"
)

func (c *Client) stateVolume(ctx context.Context, id string) (*docker.Volume, error) {
	return docker.CreateVolume(ctx, c.Client, docker.CreateVolumeOpts{
		Name: fmt.Sprintf("scribe-state-%s", id),
	})
}
