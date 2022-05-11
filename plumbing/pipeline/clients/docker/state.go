package docker

import (
	"context"
	"fmt"

	"github.com/grafana/shipwright/docker"
)

func (c *Client) stateVolume(ctx context.Context, id string) (*docker.Volume, error) {
	return docker.CreateVolume(ctx, c.Client, docker.CreateVolumeOpts{
		Name: fmt.Sprintf("shipwright-state-%s", id),
	})
}
