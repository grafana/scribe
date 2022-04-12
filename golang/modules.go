package golang

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func ModDownload() pipeline.Action {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
