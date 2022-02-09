package yarn

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

func NewStep(args ...string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}
