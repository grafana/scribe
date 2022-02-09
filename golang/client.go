package golang

import (
	"io"

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

func (c Client) BuildStep(pkg, output string) types.Step {
	return types.NewStep(func(opts types.ActionOpts) error {
		return Build(BuildOpts{
			Pkg:    pkg,
			Output: output,
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		})
	})
}

type BuildOpts struct {
	Pkg    string
	Output string
	Module string

	Stdout io.ReadWriter
	Stderr io.ReadWriter
}

func Build(opts BuildOpts) error {
	return exec.RunCommandAt(opts.Stdout, opts.Stderr, opts.Module, "go", "build", "-o", opts.Output, opts.Pkg)
}
