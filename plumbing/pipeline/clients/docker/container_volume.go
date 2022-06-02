package docker

import (
	"os"

	"github.com/docker/docker/api/types/mount"
	"github.com/grafana/scribe/docker"
)

func DefaultMounts(v *docker.Volume) ([]mount.Mount, error) {
	return []mount.Mount{
		{
			Type:   mount.TypeVolume,
			Source: v.Name,
			Target: "/opt/scribe",
			TmpfsOptions: &mount.TmpfsOptions{
				Mode: os.FileMode(0777),
			},
		},
	}, nil
}
