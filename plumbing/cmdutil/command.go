package cmdutil

import (
	"fmt"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

// StepCommand returns the command string for running a single step.
// The path argument can be omitted, which is particularly helpful if the current directory is a pipeline.
type CommandOpts struct {
	Path    string
	Step    pipeline.Step
	BuildID string
}

func StepCommand(c pipeline.Configurer, opts CommandOpts) ([]string, error) {
	args := []string{}

	for _, arg := range opts.Step.Arguments {
		if arg.Type != pipeline.ArgumentTypeString {
			continue
		}

		value, err := c.Value(arg)
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("-arg=%s=%s", arg.Key, value))
	}

	if opts.BuildID != "" {
		args = append(args, fmt.Sprintf("-build-id=%s", opts.BuildID))
	}

	cmd := append([]string{"shipwright", fmt.Sprintf("-step=%d", opts.Step.Serial)}, args...)
	if opts.Path != "" {
		cmd = append(cmd, opts.Path)
	}

	return cmd, nil
}
