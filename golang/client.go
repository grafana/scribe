package golang

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Client struct {
	Modules ModulesClient
}

func (c Client) Test() types.Step {
	return types.NewStep(func() error {
		return nil
	})
}

func (c Client) Build() types.Step {
	return types.NewStep(func() error {
		return nil
	})
}
