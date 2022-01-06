// Package main contains the CLI logic for the `shipwright` command
// The shipwright command's main responsibility is to run a pipeline.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
)

type cmdFunc func(context.Context, []string)

type arguments struct {
	path string
	cmd  cmdFunc
}

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
}

func main() {
	var (
		ctx = context.Background()
	)

	args := mustCLIArgs(os.Args[1:])

	args.cmd(ctx, os.Args[1:])
}

var examples = `Examples:
  shipwright # Runs the pipeline located in $PWD
  shipwright -path=path/to/pipeline # Runs the pipeline located in path/to/pipeline
		`

func usage(f *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(f.Output(), "Usage of shipwright: shipwright [-arg=...] [run|config]")
		f.PrintDefaults()
		fmt.Fprintln(f.Output(), examples)
		if f.ErrorHandling() == flag.ExitOnError {
			os.Exit(1)
		}
		if f.ErrorHandling() == flag.PanicOnError {
			panic("invalid argument(s)")
		}
	}
}

func mustCLIArgs(args []string) *arguments {
	f := flag.NewFlagSet("shipwright CLI", flag.ContinueOnError)
	f.Usage = usage(f)

	var (
		path string
	)

	// Here is where we define our global flags
	f.StringVar(&path, "path", ".", "Path to 'main' package that contains the shipwright pipeline")

	// Ignore Parse errors and silence the output as we want to allow extra arguments to be passed to different parts of the program
	f.SetOutput(io.Discard)

	f.Parse(args)

	var (
		action = f.Arg(0)
		cmd    cmdFunc
	)

	if f.NArg() > 1 {
		log.Println("Invalid number of arguments")
		f.Usage()
	}

	switch action {
	case "", "run":
		cmd = commands.Run
	default:
		log.Println("Unrecognized argument")
		f.Usage()
	}

	return &arguments{
		cmd: cmd,
	}
}
