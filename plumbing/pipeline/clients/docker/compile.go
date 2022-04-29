package docker

import (
	"context"

	"github.com/docker/docker/client"
)

// NewCompilePipelineContainer creates a new container to compile the Shipwright pipeline.
func NewCompilePipelineContainer(ctx context.Context, client client.APIClient) (*Container, error) {
	return nil, nil
}
