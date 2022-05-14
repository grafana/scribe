package docker

import (
	"context"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type Volume struct {
	types.Volume
}

type CreateVolumeOpts struct {
	Name string
}

func CreateVolume(ctx context.Context, cli client.APIClient, opts CreateVolumeOpts) (*Volume, error) {
	res, err := cli.VolumeCreate(ctx, volume.VolumeCreateBody{
		Name: opts.Name,
	})
	if err != nil {
		return nil, err
	}

	return &Volume{res}, nil
}

func DeleteVolume(ctx context.Context, cli client.APIClient, volume *Volume) error {
	return cli.VolumeRemove(ctx, volume.Name, false)
}

func (v *Volume) MountAt(path string, mode os.FileMode) mount.Mount {
	return mount.Mount{
		Type:   mount.TypeVolume,
		Source: v.Name,
		Target: path,
		TmpfsOptions: &mount.TmpfsOptions{
			Mode: mode,
		},
	}
}
