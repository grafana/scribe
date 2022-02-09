package golang

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

type ModulesClient struct{}

func (m *ModulesClient) Download() pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}
