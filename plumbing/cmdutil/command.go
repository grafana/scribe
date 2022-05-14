package cmdutil

import (
	"errors"
	"fmt"

	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

// CommandOpts is a list of arguments that can be provided to the StepCommand function.
type CommandOpts struct {
	// Step is the pipeline step this command is being generated for. The step contains a lot of necessary information for generating a command, mostly around arguments.
	Step pipeline.Step[pipeline.Action]

	// CompiledPipeline is an optional argument. If it is supplied, this value will be used as the first argument in the command instead of the shipwright command.
	// This option is useful scenarios where the 'shipwright' command will not be available, but the pipeline has been compiled.
	CompiledPipeline string
	// Path is an optional argument that refers to the path of the pipeline. For example, if our plan is to have this function generate `shipwright ./ci`, the 'Path' would be './ci'.
	Path string
	// BuildID is an optional argument that will be supplied to the 'shipwright' command as '-build-id'.
	BuildID string
	// State is an optional argument that is supplied as '-state'. It is a path to the JSON state file which allows steps to share data.
	State string
	// StateArgs pre-populate the state for a specific step. These strings can include references to environment variables using $.
	StateArgs map[string]string
}

// StepCommand returns the command string for running a single step.
// The path argument can be omitted, which is particularly helpful if the current directory is a pipeline.
func StepCommand(c pipeline.Configurer, opts CommandOpts) ([]string, error) {
	args := []string{}

	for _, arg := range opts.Step.Arguments {
		if arg.Type != pipeline.ArgumentTypeString && arg.Type != pipeline.ArgumentTypeSecret {
			continue
		}

		value, err := c.Value(arg)
		if err != nil {
			// If it wasn't found by the Configurer, then it's likely going to be provided in the state by another step.
			// TODO: we could actually check this by searching the pipeline.
			if errors.Is(err, plumbing.ErrorMissingArgument) {
				continue
			}
			return nil, err
		}

		args = append(args, fmt.Sprintf("-arg=%s=%s", arg.Key, value))
	}

	if opts.BuildID != "" {
		args = append(args, fmt.Sprintf("-build-id=%s", opts.BuildID))
	}

	if opts.State != "" {
		args = append(args, fmt.Sprintf("-state=%s", opts.State))
	}

	if len(opts.StateArgs) != 0 {
		for k, v := range opts.StateArgs {
			args = append(args, "-state-%s=%s", k, v)
		}
	}

	name := "shipwright"

	if p := opts.CompiledPipeline; p != "" {
		name = p
	}

	cmd := append([]string{name, fmt.Sprintf("-step=%d", opts.Step.Serial)}, args...)
	if opts.Path != "" {
		cmd = append(cmd, opts.Path)
	}

	return cmd, nil
}
