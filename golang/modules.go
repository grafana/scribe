package golang

import (
	"context"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func ModDownload() pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
