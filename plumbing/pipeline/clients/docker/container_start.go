package docker

import (
	"context"
	"io"
)

type RunContainerOpts struct {
	Container *Container
	Stdin     io.Writer
	Stdout    io.Writer
}

func StartContainer(ctx context.Context, client ContainerClient, opts RunContainerOpts) error {
	return nil
}
