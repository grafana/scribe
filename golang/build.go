package golang

import (
	"context"

	"pkg.grafana.com/shipwright/v1/golang/x"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func BuildStep(pkg, output string) pipeline.Step {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		return x.Build(ctx, x.BuildOpts{
			Pkg:    pkg,
			Output: output,
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		})
	})
}
