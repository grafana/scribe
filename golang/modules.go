package golang

import (
	"context"

	"github.com/grafana/scribe/pipeline"
)

func ModDownload() pipeline.Action {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
