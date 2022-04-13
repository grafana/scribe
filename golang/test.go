package golang

import (
	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/exec"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

func Test(sw *shipwright.Shipwright[pipeline.Action], pkg string) pipeline.Step[pipeline.Action] {
	return pipeline.NewStep(exec.Run("go", "test", pkg)).
		WithImage(plumbing.SubImage("go", sw.Opts.Version)).
		WithArguments(pipeline.ArgumentSourceFS)
}
