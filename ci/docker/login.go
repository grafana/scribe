package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	swdocker "github.com/grafana/shipwright/docker"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

var (
	ArgumentDockerAuthToken = pipeline.NewFSArgument("docker-auth-token")
)

func Login(username, password, registry pipeline.Argument) pipeline.Step[pipeline.Action] {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		u, err := opts.State.Get(username.Key)
		if err != nil {
			return err
		}

		p, err := opts.State.Get(password.Key)
		if err != nil {
			return err
		}

		r, err := opts.State.Get(registry.Key)
		if err != nil {
			return err
		}

		token, err := swdocker.Login(ctx, types.AuthConfig{
			Username:      u,
			Password:      p,
			ServerAddress: r,
		})

		return opts.State.Set(ArgumentDockerAuthToken.Key, token)
	}).
		WithArguments(username, password, registry, pipeline.ArgumentDockerSocketFS).
		Provides(ArgumentDockerAuthToken)
}
