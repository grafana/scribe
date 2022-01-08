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
func Run(ctx context.Context, stdout io.Writer, stderr io.Writer, args *plumbing.Arguments) error {
	var (
		cmdArgs = []string{"run", args.Path, "-mode", "cli"}
	)

	plog.Infoln("Running shipwright pipeline with args", cmdArgs)

	if args.Step != nil {
		cmdArgs = append(cmdArgs, "-step", strconv.Itoa(*args.Step))
	}

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func MustRunStdout(ctx context.Context, args *plumbing.Arguments) {
	if err := Run(ctx, os.Stdout, os.Stderr, args); err != nil {
		plog.Fatalln("Error running pipeline:", err)
	}
}
