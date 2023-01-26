package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

var (
	ArgumentCompiledBackend  = state.NewFileArgument("compiled-backend")
	ArgumentCompiledFrontend = state.NewDirectoryArgument("compiled-frontend")
)

func actionBuildBackend(ctx context.Context, opts pipeline.ActionOpts) error {
	f, err := os.CreateTemp("", "*")
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := opts.State.SetFileReader(ctx, ArgumentCompiledBackend, f); err != nil {
		return err
	}

	return nil
}

func actionBuildFrontend(ctx context.Context, opts pipeline.ActionOpts) error {
	path := filepath.Join(os.TempDir(), "frontend")
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	return opts.State.SetDirectory(ctx, ArgumentCompiledFrontend, path)
}

var stepBuildBackend = pipeline.NamedStep("build-backend", actionBuildBackend).
	Requires(ArgumentNodeDependencies, pipeline.ArgumentSourceFS).
	Provides(ArgumentCompiledBackend)

var stepBuildFrontend = pipeline.NamedStep("build-backend", actionBuildFrontend).
	Requires(ArgumentGoDependencies, pipeline.ArgumentSourceFS).
	Provides(ArgumentCompiledFrontend)

var PipelineBuild = Pipeline{
	Name:     "build",
	Provides: []state.Argument{ArgumentCompiledBackend, ArgumentCompiledFrontend},
	Requires: []state.Argument{ArgumentGoDependencies, ArgumentNodeDependencies},
	Steps:    []pipeline.Step{stepBuildBackend, stepBuildFrontend},
}
