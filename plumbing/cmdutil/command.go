package cmdutil

import (
	"fmt"

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
}

// StepCommand returns the command string for running a single step.
// The path argument can be omitted, which is particularly helpful if the current directory is a pipeline.
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
