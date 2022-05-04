package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type RunContainerOpts struct {
	Container *Container
	Stdout    io.Writer
	Stderr    io.Writer
}

// RunContainer starts the docker container
func RunContainer(ctx context.Context, cli client.APIClient, opts RunContainerOpts) error {
	// ContainerStart runs the container without attaching
	if err := cli.ContainerStart(ctx, opts.Container.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	var status container.ContainerWaitOKBody
	// ContainerWait will submit on the returned channels whenever the container is done.
	statusCh, errCh := cli.ContainerWait(ctx, opts.Container.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case s := <-statusCh:
		status = s
	}

	// ContainerLogs returns the stdout / stderr of the container, stopped or not.
	out, err := cli.ContainerLogs(ctx, opts.Container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})

	if err != nil {
		return err
	}

	stdcopy.StdCopy(opts.Stdout, opts.Stderr, out)

	if status.Error != nil {
		return fmt.Errorf("error waiting for container. Exit Code: '%d'. Error: %s", status.StatusCode, status.Error)
	}

	if status.StatusCode != 0 {
		return fmt.Errorf("container exited with code '%d'", status.StatusCode)
	}

	return nil
}
