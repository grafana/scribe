package yarn

import "pkg.grafana.com/shipwright/v1/plumbing/types"

func NewStep(args ...string) types.StepAction {
	return func(types.ActionOpts) error {
		return nil
	}
}
