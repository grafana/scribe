package golang

import (
	"pkg.grafana.com/shipwright/v1/exec"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type Client struct {
	Modules ModulesClient
	Opts    *types.CommonOpts
}

func New(o *types.CommonOpts) Client {
	return Client{
		Opts: o,
	}
}

func (c Client) Test(pkg string) types.Step {
	return types.NewStep(exec.Run("go", "test", pkg)).
		WithImage(plumbing.SubImage("go", c.Opts.Version)).
		WithArguments(types.ArgumentSourceFS)
}

func (c Client) Build(pkg, output string) types.Step {
	return types.NewStep(func(types.ActionOpts) error {
		return nil
	})
}
