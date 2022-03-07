package yarn

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func NewStep(args ...string) pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
