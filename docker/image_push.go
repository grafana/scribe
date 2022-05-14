package docker

import (
	"context"
	"errors"
	"io"

	"github.com/docker/docker/api/types"
)

type PushOpts struct {
	Name      string
	Registry  string
	AuthToken string
	Stdout    io.Writer
}

func Push(ctx context.Context, opts PushOpts) error {
	client := dockerClient()

	cfg, err := DefaultConfig()
	if !errors.Is(err, ErrorNoDockerConfig) {
		return err
	}

	auth := opts.AuthToken

	if opts.AuthToken == "" && cfg != nil {
		a, err := cfg.RegistryAuth(opts.Registry)
		if err != nil {
			return err
		}

		auth = a
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
