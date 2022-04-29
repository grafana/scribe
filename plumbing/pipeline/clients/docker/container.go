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
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

type Container struct {
	ID string
}

type CreateContainerOpts struct {
	Name    string
	Image   string
	Command []string
	Network *Network
	Mounts  []mount.Mount
	Env     []string
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

	containerConfig := &container.Config{
		Image: opts.Image,
		Cmd:   opts.Command,
		Env:   opts.Env,
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
		ID: res.ID,
	}, nil
}

func DeleteContainer(ctx context.Context, cli client.APIClient, container *Container) error {
	return cli.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   true,
	})
}

type CreateStepContainerOpts struct {
	Configurer pipeline.Configurer
	Step       pipeline.Step[pipeline.Action]
	Network    *Network
	Volume     *Volume
	Binary     string
	Pipeline   string
	BuildID    string
}

func CreateStepContainer(ctx context.Context, cli client.APIClient, opts CreateStepContainerOpts) (*Container, error) {
	cmd, err := cmdutil.StepCommand(opts.Configurer, cmdutil.CommandOpts{
		CompiledPipeline: opts.Binary,
		Path:             opts.Pipeline,
		Step:             opts.Step,
		BuildID:          opts.BuildID,
	})

	if err != nil {
		return nil, err
	}

	mounts, err := DefaultMounts(opts.Volume)
	if err != nil {
		return nil, err
	}

	return CreateContainer(ctx, cli, CreateContainerOpts{
		Name:    opts.Step.Name,
		Network: opts.Network,
		Mounts:  mounts,
		Command: cmd,
	})
}
