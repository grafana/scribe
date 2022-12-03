package cmdutil

import (
	"fmt"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/pipeline"
)

// CommandOpts is a list of arguments that can be provided to the StepCommand function.
type CommandOpts struct {
	args.PipelineArgs

	// Step is the pipeline step this command is being generated for. The step contains a lot of necessary information for generating a command, mostly around arguments.
	Step             pipeline.Step
	CompiledPipeline string
}

// StepCommand returns the command string for running a single step.
// The path argument can be omitted, which is particularly helpful if the current directory is a pipeline.
func StepCommand(opts CommandOpts) ([]string, error) {
	args := []string{"--client", "cli"}

	if opts.BuildID != "" {
		args = append(args, fmt.Sprintf("--build-id=%s", opts.BuildID))
	}

	if opts.State != "" {
		args = append(args, fmt.Sprintf("--state=%s", opts.State))
	}

	if opts.LogLevel != 0 {
		args = append(args, fmt.Sprintf("--log-level=%s", opts.LogLevel.String()))
	}

	if opts.Version != "" {
		args = append(args, fmt.Sprintf("--version=%s", opts.Version))
	}

	if len(opts.ArgMap) != 0 {
		for k, v := range opts.ArgMap {
			args = append(args, fmt.Sprintf("--arg=%s=%s", k, v))
		}
	}

	name := "scribe"

	if p := opts.CompiledPipeline; p != "" {
		name = p
	}

	cmd := append([]string{name, fmt.Sprintf("--step=%d", opts.Step.ID)}, args...)
	if opts.Path != "" {
		cmd = append(cmd, opts.Path)
	}

	return cmd, nil
}

type PipelineCommandOpts struct {
	CommandOpts
	Pipeline pipeline.Pipeline
}

func PipelineCommand(opts PipelineCommandOpts) ([]string, error) {
	args := []string{}

	if opts.BuildID != "" {
		args = append(args, fmt.Sprintf("--build-id=%s", opts.BuildID))
	}

	if opts.State != "" {
		args = append(args, fmt.Sprintf("--state=%s", opts.State))
	}

	if opts.LogLevel != 0 {
		args = append(args, fmt.Sprintf("--log-level=%s", opts.LogLevel))
	}

	if opts.Version != "" {
		args = append(args, fmt.Sprintf("--version=%s", opts.Version))
	}

	if opts.Event != "" {
		args = append(args, fmt.Sprintf("--event=%s", opts.Event))
	}

	if len(opts.ArgMap) != 0 {
		for k, v := range opts.ArgMap {
			args = append(args, fmt.Sprintf("--arg=%s=%s", k, v))
		}
	}

	name := "scribe"

	if p := opts.CompiledPipeline; p != "" {
		name = p
	}

	cmd := append([]string{name, fmt.Sprintf("--pipeline=%s", opts.Pipeline.Name)}, args...)
	if opts.Path != "" {
		cmd = append(cmd, opts.Path)
	}

	return cmd, nil

}
