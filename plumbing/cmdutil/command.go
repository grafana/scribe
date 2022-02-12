package cmdutil

import (
	"fmt"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

// StepCommand returns the command string for running a single step.
// The path argument can be omitted, which is particularly helpful if the current directory is a pipeline.
func StepCommand(c pipeline.Configurer, path string, step pipeline.Step) ([]string, error) {
	args := []string{}

	for _, arg := range step.Arguments {
		if arg.Type != pipeline.ArgumentTypeString {
			continue
		}

		value, err := c.Value(arg)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-arg=%s=%s", arg.Key, value))
	}

	cmd := append([]string{"shipwright", fmt.Sprintf("-step=%d", step.Serial)}, args...)
	if path != "" {
		cmd = append(cmd, path)
	}

	return cmd, nil
}
