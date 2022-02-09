package exec

import (
	"io"
	"os/exec"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

// RunCommandAt runs a given command and set of arguments at the given location
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommandAt(stdout, stderr io.Writer, path string, name string, arg ...string) error {
	plog.Infof("[%s] Running command: '%s %v'", path, name, arg)
	c := exec.Command(name, arg...)
	c.Stdout = stdout
	c.Stderr = stderr

	c.Dir = path

	return c.Run()
}

// RunCommand runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommand(stdout, stderr io.Writer, name string, arg ...string) error {
	return RunCommandAt(stdout, stderr, ".", name, arg...)
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func Run(name string, arg ...string) pipeline.StepAction {
	return func(opts pipeline.ActionOpts) error {
		return RunCommand(opts.Stdout, opts.Stderr, name, arg...)
	}
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunAt(path string, name string, arg ...string) pipeline.StepAction {
	return func(opts pipeline.ActionOpts) error {
		return RunCommandAt(opts.Stdout, opts.Stderr, path, name, arg...)
	}
}
