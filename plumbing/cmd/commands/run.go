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

// Run handles the default shipwright command, "shipwright run".
// The run command attempts to run the pipeline by using "go run ...".
// This function will exit the program if it encounters an error.
func Run(ctx context.Context, path string, stdout io.Writer, stderr io.Writer, args *plumbing.Arguments) error {
	var (
		// This will run a weird looking command, like this:
		//   go run ./demo/basic -mode drone -path ./demo/basic
		// But it's important to note that a lot happens before it actually reaches the pipeline code and produces a command like this:
		//   /tmp/random-string -mode drone -path ./demo/basic
		// So the path to the pipeline is not preserved, which is why we have to provide the path as an argument
		cmdArgs = []string{"run", path, "-mode", args.Mode.String(), "-path", args.Path}
	)

	plog.Infoln("Running shipwright pipeline with args", cmdArgs)

	if args.Step != nil {
		cmdArgs = append(cmdArgs, "-step", strconv.Itoa(*args.Step))
	}

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func MustRunStdout(ctx context.Context, path string, args *plumbing.Arguments) {
	if err := Run(ctx, path, os.Stdout, os.Stderr, args); err != nil {
		plog.Fatalln("Error running pipeline:", err)
	}
}
