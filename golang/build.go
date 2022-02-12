package golang

import (
	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/exec"
	"pkg.grafana.com/shipwright/v1/golang/x"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func Test(sw shipwright.Shipwright, pkg string) pipeline.Step {
	return pipeline.NewStep(exec.Run("go", "test", pkg)).
		WithImage(plumbing.SubImage("go", sw.Opts.Version)).
		WithArguments(pipeline.ArgumentSourceFS)
}

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
