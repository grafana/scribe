package yarn

import (
	"context"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func NewStep(args ...string) pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
