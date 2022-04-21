package docker

import "context"

type Network struct {
}

type CreateNetworkOpts struct {
}

func CreateNetwork(ctx context.Context, client ContainerClient, opts CreateNetworkOpts) (*Network, error) {
	return nil, nil
}
