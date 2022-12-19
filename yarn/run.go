package yarn

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/grafana/scribe/exec"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

var (
	ArgumentYarnCache = state.NewDirectoryArgument("yarn-cache-dir")
)

func InstallAction() pipeline.Action {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		if err := exec.Run(ctx, opts, "yarn", "install"); err != nil {
			return err
		}

		return opts.State.SetDirectory(ctx, ArgumentYarnCache, filepath.Join(".yarn", "cache"))
	}
}

func InstallStep() pipeline.Step {
	return pipeline.
		NewStep(InstallAction()).
		WithName("yarn install").
		Requires(pipeline.ArgumentSourceFS).
		Provides(ArgumentYarnCache)
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
		Requires(pipeline.ArgumentSourceFS)
}
