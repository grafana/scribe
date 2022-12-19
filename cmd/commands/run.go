package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/plog"
)

type RunOpts struct {
	// Version specifies the scribe version to run.
	// This value is used in generating the scribe image. A value will be provided if using the scribe CLI.
	// If no value is provided, then it will be replaced with "latest".
	Version string

	// Path specifies the path to the scribe pipeline.
	// This value is assumed to be "." if not provided.
	// This is not the same as the "Path" argument for the pipeline itself, which is required and used for code / config generation.
	Path string

	State string
	// Stdout is the stdout stream of the "go run" command that runs the pipeline
	// If it is not provided, it defaults to "os.Stdout"
	Stdout io.Writer
	// Stderr is the stderr stream of the "go run" command that runs the pipeline.
	// The stderr stream contains mostly logging info and is particularly useful if a problem is encountered.
	// If it is not provided, it defaults to "os.Stderr"
	Stderr io.Writer
	// Stdin is the stdin stream of the "go run" command that runs the pipeline.
	// The stdin stream is used to accept arguments in docker or cli client that were not provided in command-line arguments.
	// If it is not provided, it defaults to "os.Stdin"
	Stdin io.Reader

	// Args are arguments that are passed to the scribe pipeline
	Args *args.PipelineArgs
}

// Run handles the default scribe command, "scribe run".
// The run command attempts to run the pipeline by using "go run ...".
// This function will exit the program if it encounters an error.
// TODO: there is a function in `cmdutil` that should be able to create this command to run.
func Run(ctx context.Context, opts *RunOpts) *exec.Cmd {
	var (
		path  = opts.Path
		args  = opts.Args
		state = opts.State

		stdout  = opts.Stdout
		stderr  = opts.Stderr
		stdin   = opts.Stdin
		version = opts.Version
	)

	if stdout == nil {
		stdout = os.Stdout
	}

	if stderr == nil {
		stderr = os.Stderr
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	if version == "" {
		version = "latest"
	}

	logger := plog.New(opts.Args.LogLevel)

	// This will run a weird looking command, like this:
	//   go run ./demo/basic -client drone -path ./demo/basic
	// But it's important to note that a lot happens before it actually reaches the pipeline code and produces a command like this:
	//   /tmp/random-string -client drone -path ./demo/basic
	// So the path to the pipeline is not preserved, which is why we have to provide the path as an argument
	cmdArgs := []string{"run", path, "--client", args.Client, "--log-level", args.LogLevel.String(), "--path", args.Path, "--version", version, "--build-id", args.BuildID, "--event", args.Event, "--state", state}

	for k, v := range args.ArgMap {
		cmdArgs = append(cmdArgs, "--arg", fmt.Sprintf("%s=%s", k, v))
	}

	if args.PipelineName != nil {
		for _, v := range args.PipelineName {
			cmdArgs = append(cmdArgs, fmt.Sprintf("--pipeline=\"%s\"", v))
		}
	} else if args.Step != nil {
		cmdArgs = append(cmdArgs, "--step", strconv.FormatInt(*args.Step, 10))
	}

	logger.Infoln("Running scribe pipeline with command", append([]string{"go"}, cmdArgs...))

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin

	return cmd
}
