package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
)

type PushOpts struct {
	Name     string
	Registry string
	Stdout   io.Writer
}

func Push(ctx context.Context, opts PushOpts) error {
	client := dockerClient()

	cfg, err := DefaultConfig()
	if err != nil {
		return err
	}

	auth, err := cfg.RegistryAuth(opts.Registry)
	if err != nil {
		return err
	}

	res, err := client.ImagePush(ctx, opts.Name, types.ImagePushOptions{
		RegistryAuth: auth,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	return WriteImageLogs(res, opts.Stdout)
}
