package golang

import (
	"context"

	"github.com/grafana/shipwright/golang/x"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

func BuildStep(pkg, output string) pipeline.Step[pipeline.Action] {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		return x.RunBuild(ctx, x.BuildOpts{
			Pkg:    pkg,
			Output: output,
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		})
	})
}
