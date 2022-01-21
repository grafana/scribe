package golang

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Client struct {
	Modules ModulesClient
}

func (c Client) Test() types.StepAction {
	return func() error {
		return nil
	}
}
