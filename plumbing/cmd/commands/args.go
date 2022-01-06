package commands

import (
	"flag"
	"strconv"
)

// RunArgs are arguments provided in the Run command
type RunArgs struct {
	// Path is provided in every execution to the shipwright run command,
	// and contians the user-supplied location of the shipwright pipeline (or "." / "$PWD") by default.
	Path string

	// Step defines a specific step to run. Typically this is used in a generated third-party config
	// If Step is nil, then all steps are ran
	Step *int

	// Debug enables more verbose logging
	Debug bool
}

type OptionalInt struct {
	Value int
	Valid bool
}

func (o *OptionalInt) String() string {
	if o.Valid {
		return strconv.Itoa(o.Value)
	}

	return ""
}

func (o *OptionalInt) Set(v string) error {
	if v == "" {
		return nil
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return err
	}

	o.Value = i
	o.Valid = true

	return nil
}

// ParseRunArgs parses the "run" arguments from the args slice. These options are provided by the shipwright command and are typically not user-specified
func ParseRunArgs(args []string) (*RunArgs, error) {
	var (
		flagSet = flag.NewFlagSet("run", flag.ContinueOnError)
		step    OptionalInt
		path    string
		debug   bool
	)

	flagSet.StringVar(&path, "path", ".", "Path to 'main' package that contains the shipwright pipeline")
	flagSet.BoolVar(&debug, "debug", false, "Enable debug logging")
	flagSet.Var(&step, "step", "Enable debug logging")

	// Workaround: The '-mode' flag will be set as it is an option provided internally that is forwarded to the pipeline
	flagSet.String("mode", "cli", "")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	runArgs := &RunArgs{
		Path:  path,
		Debug: debug,
	}

	if step.Valid {
		runArgs.Step = &step.Value
	}

	return runArgs, nil
}

// MustParseRunArgs parses the "run" arguments from the args slice. These options are provided by the shipwright command and are typically not user-specified
func MustParseRunArgs(args []string) *RunArgs {
	v, err := ParseRunArgs(args)
	if err != nil {
		panic(err)
	}

	return v
}
