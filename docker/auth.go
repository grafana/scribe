package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
)

// Login calls the Docker `RegistryLogin` function. Apparently though this function just does absolutely nothing so don't even bother.
// Successful auths don't return a token. Failed auths don't return an error and they still just say 'Login successful'.
// Deprecated. Use at the risk of your own sanity.
func Login(ctx context.Context, cfg types.AuthConfig) (registry.AuthenticateOKBody, error) {
	client := dockerClient()
	return client.RegistryLogin(ctx, cfg)
}
