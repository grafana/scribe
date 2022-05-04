package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Network struct {
	ID string
}

type CreateNetworkOpts struct {
	Name string
}

func CreateNetwork(ctx context.Context, cli client.APIClient, opts CreateNetworkOpts) (*Network, error) {
	res, err := cli.NetworkCreate(ctx, opts.Name, types.NetworkCreate{})
	if err != nil {
		return nil, err
	}

	return &Network{
		ID: res.ID,
	}, nil
}

func DeleteNetwork(ctx context.Context, cli client.APIClient, network *Network) error {
	return cli.NetworkRemove(ctx, network.ID)
}
