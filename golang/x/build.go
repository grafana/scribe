package x

import (
	"context"
	"io"

	"github.com/grafana/shipwright/exec"
)

type BuildOpts struct {
	Env    []string
	Pkg    string
	Output string
	Module string

	Stdout io.Writer
	Stderr io.Writer
}

func Build(ctx context.Context, opts BuildOpts) error {
	return exec.RunCommandWithOpts(ctx, exec.RunOpts{
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Path:   opts.Module,
		Name:   "go",
		Args:   []string{"build", "-o", opts.Output, opts.Pkg},
		Env:    opts.Env,
	})
}
