package exec

import (
	"context"
	"io"
	"os/exec"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

// RunCommandAt runs a given command and set of arguments at the given location
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommandAt(ctx context.Context, stdout, stderr io.Writer, path string, name string, arg ...string) error {
	c := exec.CommandContext(ctx, name, arg...)
	c.Stdout = stdout
	c.Stderr = stderr

	c.Dir = path

	return c.Run()
}

// RunCommand runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommand(ctx context.Context, stdout, stderr io.Writer, name string, arg ...string) error {
	return RunCommandAt(ctx, stdout, stderr, ".", name, arg...)
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func Run(name string, arg ...string) pipeline.StepAction {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		return RunCommand(ctx, opts.Stdout, opts.Stderr, name, arg...)
	}
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunAt(path string, name string, arg ...string) pipeline.StepAction {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		return RunCommandAt(ctx, opts.Stdout, opts.Stderr, path, name, arg...)
	}
}
