package commands

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strconv"
)

// Run handles the default shipwright command, "shipwright run".
// The run command attempts to run the pipeline by using "go run ...".
// This function will exit the program if it encounters an error.
func Run(ctx context.Context, args []string) {
	var (
		runArgs = MustParseRunArgs(args)
		cmdArgs = []string{"run", runArgs.Path, "-mode", "cli"}
	)

	if runArgs.Step != nil {
		cmdArgs = append(cmdArgs, "-step", strconv.Itoa(*runArgs.Step))
	}

	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalln("Failed to run pipeline. Error:", err)
	}
}
