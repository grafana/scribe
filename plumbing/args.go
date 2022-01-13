package plumbing

import (
	"flag"
	"fmt"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

// Arguments are provided to the `shipwright` command.
type Arguments struct {
	Mode RunModeOption

	// Path is provided in every execution to the shipwright run command,
	// and contians the user-supplied location of the shipwright pipeline (or "." / "$PWD") by default.
	Path string

	// Step defines a specific step to run. Typically this is used in a generated third-party config
	// If Step is nil, then all steps are ran
	Step *int
}

func ParseArguments(args []string) (*Arguments, error) {
	var (
		flagSet      = flag.NewFlagSet("run", flag.ContinueOnError)
		step         OptionalInt
		mode         RunModeOption = RunModeCLI
		logLevel     plog.LogLevel
		pathOverride string
	)

	flagSet.Usage = usage(flagSet)

	flagSet.Var(&mode, "mode", "cli|docker|drone. Default: cli")
	flagSet.Var(&step, "step", "")
	flagSet.Var(&logLevel, "log-level", "")
	flagSet.StringVar(&pathOverride, "path", "", "Providing the path argument overrides the path provided in the CLI argument. This can be used when running the pipeline directly.")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	arguments := &Arguments{
		Mode: mode,
	}

	if step.Valid {
		arguments.Step = &step.Value
	}

	path := flagSet.Arg(flagSet.NArg() - 1)

	if path == "" {
		path = "."
	}

	if pathOverride != "" {
		path = pathOverride
	}

	arguments.Path = path
	return arguments, nil

}

var examples = `Examples:
  shipwright # Runs the pipeline located in $PWD
  shipwright path/to/pipeline # Runs the pipeline located in path/to/pipeline
  shipwright -mode=drone path/to/pipeline # Generates a Drone config using the pipeline defined at the specified path`

func usage(f *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(f.Output(), "Usage of shipwright: shipwright [-arg=...] [-mode=run|drone|docker] [path]")
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
