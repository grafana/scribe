package shipwright

import (
	"flag"
	"io"

	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
)

type Opts struct {
	Mode    RunMode
	RunArgs commands.RunArgs
}

// CommonOpts are provided in the Client's Init function, which includes options that are common to all clients, like
// logging, output, and debug options
type CommonOpts struct {
	Output io.Writer
}

// ParseCLIOpts parses the CLI opts when running a pipeline. These options are used to initialize the client.
func ParseCLIOpts(args []string) (*Opts, error) {
	// Remove the command name from the arguments
	args = args[1:]

	var (
		flagSet = flag.NewFlagSet("shipwright", flag.ContinueOnError)
		opts    = &Opts{}
	)

	flagSet.Var(&opts.Mode, "mode", "An internal flag that defines what mode the shipwright pipeline is running in.")

	flagSet.SetOutput(io.Discard)
	flagSet.Parse(args)

	mode := opts.Mode

	if err := mode.Client.Parse(args); err != nil {
		return nil, err
	}

	return opts, nil
}
