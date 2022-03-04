package commands

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strconv"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

type RunOpts struct {
	// Version specifies the shipwright version to run.
	// This value is used in generating the shipwright image. A value will be provided if using the shipwright CLI.
	// If no value is provided, then it will be replaced with "latest".
	Version string

	// Path specifies the path to the shipwright pipeline.
	// This value is assumed to be "." if not provided.
	// This is not the same as the "Path" argument for the pipeline itself, which is required and used for code / config generation.
	Path string

	// Stdout is the stdout stream of the "go run" command that runs the pipeline
	// If it is not provided, it defaults to "os.Stdout"
	Stdout io.Writer
	// Stderr is the stderr stream of the "go run" command that runs the pipeline.
	// The stderr stream contains mostly logging info and is particularly useful if a problem is encountered.
	// If it is not provided, it defaults to "os.Stderr"
	Stderr io.Writer
	// Stdin is the stdin stream of the "go run" command that runs the pipeline.
	// The stdin stream is used to accept arguments in docker or cli mode that were not provided in command-line arguments.
	// If it is not provided, it defaults to "os.Stdin"
	Stdin io.Reader

	// Args are arguments that are passed to the shipwright pipeline
	Args *plumbing.PipelineArgs
}

// Run handles the default shipwright command, "shipwright run".
// The run command attempts to run the pipeline by using "go run ...".
// This function will exit the program if it encounters an error.
// TODO: there is a function in `cmdutil` that should be able to create this command to run.
func Run(ctx context.Context, opts *RunOpts) error {
	var (
		path = opts.Path
		args = opts.Args

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
	//   go run ./demo/basic -mode drone -path ./demo/basic
	// But it's important to note that a lot happens before it actually reaches the pipeline code and produces a command like this:
	//   /tmp/random-string -mode drone -path ./demo/basic
	// So the path to the pipeline is not preserved, which is why we have to provide the path as an argument
	cmdArgs := []string{"run", path, "-mode", args.Mode.String(), "-log-level", args.LogLevel.String(), "-path", args.Path, "-version", version, "-build-id", args.BuildID}

	logger.Infoln("Running shipwright pipeline with args", cmdArgs)

	if args.Step != nil {
		cmdArgs = append(cmdArgs, "-step", strconv.Itoa(*args.Step))
	}

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func MustRun(ctx context.Context, opts *RunOpts) {
	if err := Run(ctx, opts); err != nil {
		panic(err)
	}
}
