package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

func actionInstallFrontendDeps(ctx context.Context, opts pipeline.ActionOpts) error {
	opts.Logger.Infoln("Installing frontend dependencies...")
	time.Sleep(1 * time.Second)
	opts.Logger.Infoln("Done installing frontend dependencies")

	// yarn install...
	// for demo purposes just creating an empty dir...
	path := filepath.Join(os.TempDir(), "frontend-deps")
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	b, err := exec.Command("touch", filepath.Join(path, "a")).CombinedOutput()
	opts.Logger.Infoln(string(b), "error", err)
	b, err = exec.Command("ls", "-al", path).CombinedOutput()
	opts.Logger.Infoln(string(b), "error", err)
	return opts.State.SetDirectory(ctx, ArgumentNodeDependencies, path)
}

func actionInstallBackendDeps(ctx context.Context, opts pipeline.ActionOpts) error {
	opts.Logger.Infoln("Installing backend dependencies...")
	time.Sleep(1 * time.Second)
	opts.Logger.Infoln("Done installing backend dependencies")

	path := filepath.Join(os.TempDir(), "backend-deps")
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	b, err := exec.Command("touch", filepath.Join(path, "a")).CombinedOutput()
	opts.Logger.Infoln(string(b), "error", err)
	b, err = exec.Command("ls", "-al", path).CombinedOutput()
	opts.Logger.Infoln(string(b), "error", err)
	return opts.State.SetDirectory(ctx, ArgumentGoDependencies, path)
}

var (
	ArgumentGoDependencies   = state.NewDirectoryArgument("go-dependencies")
	ArgumentNodeDependencies = state.NewDirectoryArgument("node-dependencies")
)

var stepInstallFrontendDeps = pipeline.NamedStep("install frontend dependencies", actionInstallFrontendDeps).
	Provides(ArgumentNodeDependencies).
	Requires(pipeline.ArgumentSourceFS)

var stepInstallBackendDeps = pipeline.NamedStep("install backend dependencies", actionInstallBackendDeps).
	Provides(ArgumentGoDependencies).
	Requires(pipeline.ArgumentSourceFS)

var PipelineDependencies = Pipeline{
	Name:     "dependencies",
	Provides: []state.Argument{ArgumentNodeDependencies, ArgumentGoDependencies},
	Steps: []pipeline.Step{
		stepInstallFrontendDeps,
		stepInstallBackendDeps,
	},
}

// ExtractBackendDependencies retrieves the backend dependencies from the state handler and places them in the appropriate place on the filesystem.
// This function assumes that your step requires the "ArgumentGoDependencies" argument
func ExtractBackendDependencies(st state.Handler) error {
	return nil
}

// ExtractBackendDependencies retrieves the backend dependencies from the state handler and places them in the appropriate place on the filesystem.
// This function assumes that your step requires the "ArgumentGoDependencies" argument
func ExtractFrontendDependencies(st state.Handler) error {
	return nil
}
