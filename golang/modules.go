package golang

import "pkg.grafana.com/shipwright/v1/types"

type ModulesClient struct{}

func (m *ModulesClient) Download() types.Step {
	return func() error {
		return nil
	}
}
