package makefile

import (
	"context"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func Target(name string) pipeline.StepAction {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		return nil
	}
}
