package docker

import (
	"context"

	"github.com/docker/docker/api/types"
)

func Login(ctx context.Context, cfg types.AuthConfig) (string, error) {
	client := dockerClient()
	res, err := client.RegistryLogin(ctx, cfg)
	if err != nil {
		return "", err
	}

	return res.IdentityToken, nil
}
