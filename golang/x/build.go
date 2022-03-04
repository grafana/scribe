package x

import (
	"context"
	"io"

	"pkg.grafana.com/shipwright/v1/exec"
)

type BuildOpts struct {
	Pkg    string
	Output string
	Module string

	Stdout io.ReadWriter
	Stderr io.ReadWriter
}

func Build(ctx context.Context, opts BuildOpts) error {
	return exec.RunCommandAt(ctx, opts.Stdout, opts.Stderr, opts.Module, "go", "build", "-o", opts.Output, opts.Pkg)
}