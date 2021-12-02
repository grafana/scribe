package yarn

import "pkg.grafana.com/shipwright/v1/types"

func NewStep(args ...string) types.Step {
	return func() error {
		return nil
	}
}
