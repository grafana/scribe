package golang

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func ModDownload() pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
