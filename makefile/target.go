package makefile

import (
	"context"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func Target(name string) pipeline.Action {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		return nil
	}
}
