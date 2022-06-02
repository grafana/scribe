package golang

import (
	"context"

	"github.com/grafana/scribe/golang/x"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func BuildStep(pkg, output string) pipeline.Step {
	return pipeline.NewStep(func(ctx context.Context, opts pipeline.ActionOpts) error {
		return x.RunBuild(ctx, x.BuildOpts{
			Pkg:    pkg,
			Output: output,
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		})
	})
}
