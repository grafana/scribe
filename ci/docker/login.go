package docker

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/grafana/scribe/plumbing/pipeline"
)

var (
	ArgumentDockerAuthToken = pipeline.NewStringArgument("docker-auth-token")
)

func Login(username, password pipeline.Argument) pipeline.Step {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		u, err := opts.State.GetString(username)
		if err != nil {
			return err
		}

		p, err := opts.State.GetString(password)
		if err != nil {
			return err
		}

		auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, u, p)))

		// Bro this function (more specifically the one in 'github.com/docker/docker' literally doesn't do ANYTHING.
		// res, err := swdocker.Login(ctx, types.AuthConfig{
		// 	Auth:          auth,
		// 	ServerAddress: r,
		// })

		// if err != nil {
		// 	return err
		// }

		return opts.State.SetString(ArgumentDockerAuthToken, auth)
	}).
		WithArguments(username, password, pipeline.ArgumentDockerSocketFS).
		Provides(ArgumentDockerAuthToken)
}
