package yarn

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func NewStep(args ...string) pipeline.Action {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
