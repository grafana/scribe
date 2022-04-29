package docker

import (
	"context"
	"fmt"
	"io"
	"sync"

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
		return fmt.Errorf("Error waiting for container. Code: '%d'. Error: %s", status.StatusCode, status.Error)
	}

	return nil
}

// ContainerWaitGroup takes a list of containers added via the Add function and runs them in parallel using the Wait() function.
type ContainerWaitGroup struct {
	Containers []RunContainerOpts

	wg *sync.WaitGroup
}

// Add adds a new container to the list of parallel-ran containers.
// For 'stdout' and 'stderr' in the provided RunContainerOpts, it's suggested to create a new sub-logger using
// (logrus.FieldLogger).WithFields.
func (w *ContainerWaitGroup) Add(c RunContainerOpts) {
	w.Containers = append(w.Containers, c)
}

// Wait will block the current goroutine until the Containers provided via Add have completed.
// If an error is returned by a container or the context is cancelled then the function will immediately stop.
func (w *ContainerWaitGroup) Wait(ctx context.Context, cli client.APIClient) error {
	var (
		doneChan = make(chan bool)
		errChan  = make(chan error)
	)

	w.wg.Add(len(w.Containers))

	for _, container := range w.Containers {
		go func(container RunContainerOpts) {
			if err := RunContainer(ctx, cli, container); err != nil {
				errChan <- err
			}

			w.wg.Done()
		}(container)
	}

	select {
	case <-ctx.Done():
		return context.Canceled
	case <-doneChan:
		return nil
	case err := <-errChan:
		return fmt.Errorf("error encountered running docker container: %w", err)
	}
}

// NewContainerWaitGroup initializes a ContainerWaitGroup with an empty list and an empty waitgroup.
// This function should be used instead of a &ContainerWaitGroup{} literal.
func NewContainerWaitGroup() *ContainerWaitGroup {
	return &ContainerWaitGroup{
		Containers: []RunContainerOpts{},
		wg:         &sync.WaitGroup{},
	}
}
