package plumbing

import (
	"flag"
	"fmt"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

// PipelineArgs are provided to the `shipwright` command.
type PipelineArgs struct {
	Mode RunModeOption

	// Path is provided in every execution to the shipwright run command,
	// and contians the user-supplied location of the shipwright pipeline (or "." / "$PWD") by default.
	Path    string
	Version string

	// Step defines a specific step to run. Typically this is used in a generated third-party config
	// If Step is nil, then all steps are ran
	Step *int

	// CanStdinPrompt is true if the pipeline can prompt for absent arguments via stdin
	CanStdinPrompt bool

	// ArgMap is a map populated by arguments provided using the `-arg` flag.
	// Example usage: `-arg={key}={value}
	ArgMap ArgMap

	// LogLvel defines how detailed the output logs in the pipeline should be.
	// Possible options are [debug, info, warn, error].
	// The default value is warn.
	LogLevel plog.LogLevel
}

func ParseArguments(args []string) (*PipelineArgs, error) {
	var (
		flagSet                     = flag.NewFlagSet("run", flag.ContinueOnError)
		mode          RunModeOption = RunModeCLI
		logLevel                    = plog.LogLevelWarn
		step          OptionalInt
		pathOverride  string
		version       string
		noStdinPrompt bool
		argMap        = ArgMap(map[string]string{})
	)

	flagSet.Usage = usage(flagSet)

	flagSet.Var(&mode, "mode", "cli|docker|drone. Default: cli")
	flagSet.Var(&step, "step", "A number that defines what specific step to run")
	flagSet.Var(&logLevel, "log-level", "The level of detail in the pipeline's log output. Default: 'warn'. Options: [debug, info, warn, error]")
	flagSet.Var(&argMap, "arg", "")
	flagSet.BoolVar(&noStdinPrompt, "no-stdin", false, "If this flag is provided, then the CLI pipeline will not request absent arguments via stdin")
	flagSet.StringVar(&pathOverride, "path", "", "Providing the path argument overrides the $PWD of the pipeline for generation")
	flagSet.StringVar(&version, "version", "latest", "The version is provided by the 'shipwright' command, however if only using 'go run', it can be provided here")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	arguments := &PipelineArgs{
		CanStdinPrompt: !noStdinPrompt,
		Mode:           mode,
		Version:        version,
		LogLevel:       logLevel,
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
	arguments.ArgMap = argMap
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
