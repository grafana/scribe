package x

import (
	"context"
	"io"
	"os/exec"

	swexec "github.com/grafana/shipwright/exec"
)

type BuildOpts struct {
	Env    []string
	Pkg    string
	Output string
	Module string

	Stdout io.Writer
	Stderr io.Writer
}

func Build(ctx context.Context, opts BuildOpts) *exec.Cmd {
	return swexec.CommandWithOpts(ctx, swexec.RunOpts{
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Path:   opts.Module,
		Name:   "go",
		Args:   []string{"build", "-o", opts.Output, opts.Pkg},
		Env:    opts.Env,
	})
}

func RunBuild(ctx context.Context, opts BuildOpts) error {
	return Build(ctx, opts).Run()
}
