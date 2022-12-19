package x

import (
	"context"
	"io"
	"os/exec"

	swexec "github.com/grafana/scribe/exec"
)

type BuildOpts struct {
	Env     []string
	Args    []string
	Pkg     string
	Output  string
	Module  string
	LDFlags string

	Stdout io.Writer
	Stderr io.Writer
}

func Build(ctx context.Context, opts BuildOpts) *exec.Cmd {
	// for the go build command optional arguments have to come before the -o output and package name we are building
	fullArgs := append([]string{"build"}, opts.Args...)
	fullArgs = append(fullArgs, "-o", opts.Output)
	if opts.LDFlags != "" {
		fullArgs = append(fullArgs, "-ldflags", opts.LDFlags)
	}

	fullArgs = append(fullArgs, opts.Pkg)
	return swexec.CommandWithOpts(ctx, swexec.RunOpts{
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Path:   opts.Module,
		Name:   "go",
		Args:   fullArgs,
		Env:    opts.Env,
	})
}

func RunBuild(ctx context.Context, opts BuildOpts) error {
	return Build(ctx, opts).Run()
}
