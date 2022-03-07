package x

import (
	"context"
	"io"

	"github.com/grafana/shipwright/exec"
)

type BuildOpts struct {
	Pkg    string
	Output string
	Module string

	Stdout io.Writer
	Stderr io.Writer
}

func Build(ctx context.Context, opts BuildOpts) error {
	return exec.RunCommandAt(ctx, opts.Stdout, opts.Stderr, opts.Module, "go", "build", "-o", opts.Output, opts.Pkg)
}
