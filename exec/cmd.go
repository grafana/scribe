package exec

import (
	"os"
	"os/exec"

	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// RunCommandAt runs a given command and set of arguments at the given location
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommandAt(path string, name string, arg ...string) error {
	plog.Infof("[%s] Running command: '%s %v'", path, name, arg)
	c := exec.Command(name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	c.Dir = path

	return c.Run()
}

// RunCommand runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunCommand(name string, arg ...string) error {
	return RunCommandAt(".", name, arg...)
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func Run(name string, arg ...string) types.StepAction {
	return func() error {
		return RunCommand(name, arg...)
	}
}

// Run returns an action that runs a given command and set of arguments.
// The command's stdout and stderr are assigned the systems' stdout/stderr streams.
func RunAt(path string, name string, arg ...string) types.StepAction {
	return func() error {
		return RunCommandAt(path, name, arg...)
	}
}
