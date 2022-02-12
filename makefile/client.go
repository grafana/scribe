package makefile

import (
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func Target(name string) pipeline.StepAction {
	return func(opts pipeline.ActionOpts) error {
		return nil
	}
}
