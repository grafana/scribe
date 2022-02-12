package git

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

var (
	ArgGitDescription = pipeline.NewStringArgument("git-description")
)

type DescribeOpts struct {
	Tags   bool
	Dirty  bool
	Always bool
}

func DescribeAction(opts DescribeOpts) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}

func Describe(opts DescribeOpts) pipeline.Step {
	return pipeline.NewStep(DescribeAction(opts)).Provides(ArgGitDescription)
}
