package yarn

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/grafana/scribe/exec"
	"github.com/grafana/scribe/plumbing/pipeline"
)

var (
	ArgumentYarnCache = pipeline.NewDirectoryArgument("yarn-cache-dir")
)

func InstallAction() pipeline.Action {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		if err := exec.Run(ctx, opts, "yarn", "install"); err != nil {
			return err
		}

		return opts.State.SetDirectory(ArgumentYarnCache, filepath.Join(".yarn", "cache"))
	}
}

func InstallStep() pipeline.Step {
	return pipeline.
		NewStep(InstallAction()).
		WithName("yarn install").
		Provides(ArgumentYarnCache).
		WithArguments(pipeline.ArgumentSourceFS)
}

func RunAction(script ...string) pipeline.Action {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		// For now, just run the yarn command.
		// In the future, we can verify that the script exists in the "scripts" object.
		return exec.Run(ctx, opts, "yarn", script...)
	}
}

func RunStep(script ...string) pipeline.Step {
	n := append([]string{"yarn"}, script...)
	name := strings.Join(n, " ")

	action := RunAction(script...)

	return pipeline.NewStep(action).
		WithName(name).
		WithArguments(pipeline.ArgumentSourceFS)
}
