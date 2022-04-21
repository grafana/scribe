package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type ContainerClient interface{}

type Container struct {
	ID string
}

type CreateContainerOpts struct {
	Name    string
	Image   string
	Command []string
}

// CreateContainer creates a new container using the ContainerClient.
// In order for a container to be created, the image has to exist on the machine, so this function will also attempt to pull the provided image.
func CreateContainer(ctx context.Context, cli client.APIClient, opts CreateContainerOpts) (*Container, error) {
	buf := bytes.NewBuffer(nil)
	r, err := cli.ImagePull(ctx, opts.Image, types.ImagePullOptions{})
	if err != nil {
		if r != nil {
			io.Copy(buf, r)
			err = fmt.Errorf("output: %s. Error: %w", buf.String(), err)
		}

		return nil, err
	}

	res, err := cli.ContainerCreate(ctx, &container.Config{
		Image: opts.Image,
		Cmd:   opts.Command,
	}, nil, nil, nil, "")

	if err != nil {
		return nil, err
	}

	return &Container{
		ID: res.ID,
	}, nil
}
