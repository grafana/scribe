package golang

import (
	"pkg.grafana.com/shipwright/v1/golang/x"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func BuildStep(pkg, output string) pipeline.Step {
	return pipeline.NewStep(func(opts pipeline.ActionOpts) error {
		return x.Build(x.BuildOpts{
			Pkg:    pkg,
			Output: output,
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		})
	})
}
