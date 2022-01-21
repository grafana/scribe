// Package main contains the CLI logic for the `shipwright` command
// The shipwright command's main responsibility is to run a pipeline.
package main

import (
	"context"
	"log"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}

func main() {
	var (
		ctx = context.Background()
	)

	args := commands.MustParseArgs(os.Args[1:])

	commands.MustRunStdout(ctx, args.Path, args)
}
