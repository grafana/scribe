package dagger

import (
	"context"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/grafana/scribe/pipelineutil"
)

func CompilePipeline(ctx context.Context, d *dagger.Client, src, gomod, pipeline string) (*dagger.Directory, error) {
	var (
		dir     = d.Host().Directory(src)
		builder = d.Container().From("golang:1.19").WithMountedDirectory("/src", dir)
	)

	path, err := filepath.Rel(src, gomod)
	if err != nil {
		return nil, err
	}
	cmd := pipelineutil.GoBuild(ctx, pipelineutil.GoBuildOpts{
		Pipeline: pipeline,
		Module:   path,
		Output:   "/opt/scribe/pipeline",
		LDFlags:  `-extldflags "-static"`,
	})

	builder = builder.WithEnvVariable("GOOS", "linux")
	builder = builder.WithEnvVariable("GOARCH", "amd64")
	builder = builder.WithEnvVariable("CGO_ENABLED", "0")
	builder = builder.WithWorkdir("/src")

	builder = builder.Exec(dagger.ContainerExecOpts{
		Args: cmd.Args,
	})

	return builder.Directory("/opt/scribe"), nil
}
