package git

import (
	"context"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

var (
	ArgGitDescription = state.NewStringArgument("git-description")
)

type DescribeOpts struct {
	Tags   bool
	Dirty  bool
	Always bool
}

func DescribeAction(opts DescribeOpts) pipeline.Action {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}

func Describe(opts DescribeOpts) pipeline.Step {
	return pipeline.NewStep(DescribeAction(opts)).Provides(ArgGitDescription)
}
