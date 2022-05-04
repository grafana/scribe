package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type Container struct {
	CreateOpts CreateContainerOpts
	ID         string
}

type CreateContainerOpts struct {
	Name    string
	Image   string
	Command []string
	Network *Network
	Mounts  []mount.Mount
	Env     []string
	Workdir string

	// Out defines where to write output when pulling a docker image
	Out io.Writer
}

// CreateContainer creates a new container using the ContainerClient.
// In order for a container to be created, the image has to exist on the machine, so this function will also attempt to pull the provided image.
func CreateContainer(ctx context.Context, cli client.APIClient, opts CreateContainerOpts) (*Container, error) {
	buf := bytes.NewBuffer(nil)
	r, err := cli.ImagePull(ctx, opts.Image, types.ImagePullOptions{})
	if err != nil {
		if r != nil {
			io.Copy(buf, r)
			err = fmt.Errorf("Output: %s. Error: %w", buf.String(), err)
		}

		return nil, err
	}

	if err := WriteImageLogs(r, opts.Out); err != nil {
		return nil, err
	}

	containerConfig := &container.Config{
		Image:      opts.Image,
		Cmd:        opts.Command,
		Env:        opts.Env,
		WorkingDir: opts.Workdir,
	}

	hostConfig := &container.HostConfig{
		Mounts: opts.Mounts,
	}

	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}

	if opts.Network != nil {
		hostConfig.NetworkMode = container.NetworkMode(opts.Network.ID)
		netConfig.EndpointsConfig[opts.Network.ID] = &network.EndpointSettings{}
	}

	// ContainerCreate(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *specs.Platform, containerName string) (containertypes.ContainerCreateCreatedBody, error)
	res, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, netConfig, nil, "")
	if err != nil {
		return nil, err
	}

	return &Container{
		ID:         res.ID,
		CreateOpts: opts,
	}, nil
}

func DeleteContainer(ctx context.Context, cli client.APIClient, container *Container) error {
	return cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
	})
}
