package main

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/grafana/scribe/plumbing/pipeline"
)

var (
	ArgumentDockerAuthToken = pipeline.NewStringArgument("docker-auth-token")
)

func StepDockerLogin(username, password pipeline.Argument) pipeline.Step {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		u, err := opts.State.GetString(username)
		if err != nil {
			return err
		}

		p, err := opts.State.GetString(password)
		if err != nil {
			return err
		}

		authString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{"username": "%s", "password": "%s"}`, u, p)))
		return opts.State.SetString(ArgumentDockerAuthToken, authString)
	}).
		WithArguments(username, password, pipeline.ArgumentDockerSocketFS).
		Provides(ArgumentDockerAuthToken)
}
