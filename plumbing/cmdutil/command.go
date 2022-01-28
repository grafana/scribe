package cmdutil

import (
	"fmt"

	"pkg.grafana.com/shipwright/v1/plumbing/config"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// StepCommand returns the command string for running a single step
func StepCommand(c config.Configurer, path string, step types.Step) ([]string, error) {
	args := []string{}

	for _, arg := range step.Arguments {
		if arg.Type != types.ArgumentTypeString {
			continue
		}

		value, err := c.Value(arg)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-arg=%s=%s", arg.Key, value))
	}

	cmd := append([]string{"shipwright", fmt.Sprintf("-step=%d", step.Serial)}, args...)
	cmd = append(cmd, path)

	return cmd, nil
}
