// Package main contains the CLI logic for the `shipwright` command
// The shipwright command's main responsibility is to run a pipeline.
package main

import (
	"context"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

// Arguments provided at compile-time
var (
	Version = "latest"
)

func main() {
	plog.Infoln("Running version", Version)
	var (
		ctx = context.Background()
	)

	args := commands.MustParseArgs(os.Args[1:])

	commands.MustRun(ctx, &commands.RunOpts{
		Version: Version,
		Path:    args.Path,
		Args:    args,
	})
}
