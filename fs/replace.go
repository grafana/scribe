package fs

import (
	"context"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func Replace(file string, content string) pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}

func ReplaceString(file string, content string) pipeline.StepAction {
	return func(context.Context, pipeline.ActionOpts) error {
		return nil
	}
}
