package args

import (
	"errors"
	"net/url"
	"os"
	"strings"

	flag "github.com/spf13/pflag"

	"github.com/grafana/scribe/stringutil"
	"github.com/sirupsen/logrus"
)

// PipelineArgs are provided to the `scribe` command.
type PipelineArgs struct {
	Client string

	// Path is provided in every execution to the scribe run command,
	// and contians the user-supplied location of the scribe pipeline (or "." / "$PWD") by default.
	Path    string
	Version string

	// Step defines a specific step to run. Typically this is used in a generated third-party config
	// If Step is nil, then all steps are ran
	Step *int64

	// BuildID is a unique identifier typically assigned by a CI system.
	// In Dagger / CLI clients, this will likely be populated by a random UUID if not provided.
	BuildID string

	// CanStdinPrompt is true if the pipeline can prompt for absent arguments via stdin
	CanStdinPrompt bool

	// ArgMap is a map populated by arguments provided using the `-arg` flag.
	// Example usage: `-arg={key}={value}
	ArgMap ArgMap

	// LogLevel defines how detailed the output logs in the pipeline should be.
	// Possible options are [debug, info, warn, error].
	// The default value is warn.
	LogLevel logrus.Level

	// State is a URL where the build state is stored.
	// Examples:
	// * 'fs:///var/scribe/state.json' - Uses a JSON file to store the state.
	// * 'fs:///c:/scribe/state.json' - Uses a JSON file to store the state, but on Windows.
	// * 'fs:///var/scribe/state/' - Stores the state file in the given directory, using a randomly generated ID to store the state.
	//    * This might be a good option if implementing a Scribe client in a provider.
	// * 's3://bucket-name/path'
	// * 'gcs://bucket-name/path'
	// If 'State' is not provided, then one is created using os.Tmpdir.
	State string

	// PipelineName can be provided in a multi-pipeline setup to run an entire pipeline rather than the entire suite of pipelines.
	PipelineName []string

	// Event can be provided in a multi-pipeline setup locally to simulate an event.
	Event string
}

type pipelineNames struct {
	names []string
}

func (p *pipelineNames) String() string {
	return strings.Join(p.names, ",")
}

func (p *pipelineNames) Set(value string) error {
	p.names = append(p.names, value)
	return nil
}

func (p *pipelineNames) Type() string {
	return "[]string"
}

func ParseArguments(args []string) (*PipelineArgs, error) {
	var defaultState = &url.URL{
		Scheme: "file",
		Path:   os.TempDir(),
	}

	var (
		flagSet       = flag.NewFlagSet("run", flag.ContinueOnError)
		client        string
		step          OptionalInt
		logLevel      string
		pathOverride  string
		version       string
		buildID       string
		noStdinPrompt bool
		argMap        = ArgMap(map[string]string{})
		state         string
		event         string
		pipelineName  pipelineNames
	)

	// Flags with shorthand options
	flagSet.StringVarP(&client, "client", "c", "dagger", "dagger|drone. Default: dagger")
	flagSet.StringVarP(&logLevel, "log-level", "l", "info", "The level of detail in the pipeline's log output. Default: 'warn'. Options: [trace, debug, info, warn, error]")
	flagSet.StringVarP(&buildID, "build-id", "b", stringutil.Random(12), "A unique identifier typically assigned by a build system. Defaults to a random string if no build ID is provided")
	flagSet.StringVarP(&state, "state", "s", defaultState.String(), "A URI that refers to a state file or directory where state between steps is stored. Must include a protocol, like 'file://', 'gcs://', or 's3://'")
	flagSet.StringVarP(&event, "event", "e", "git-commit", "The name of an event to run. The default behavior is to run all pipelines that do not have a source event")
	flagSet.VarP(&pipelineName, "pipeline", "p", "A pipeline name, giving a value for this flag will result in only the pipeline of the specified name being executed. The default empty string will run all pipelines.")

	flagSet.Var(&step, "step", "A number that defines what specific step to run")
	flagSet.Var(&argMap, "arg", "Provide pre-available arguments for use in pipeline steps. This argument can be provided multiple times. Format: '-arg={key}={value}")
	flagSet.BoolVar(&noStdinPrompt, "no-stdin", false, "If this flag is provided, then the CLI pipeline will not request absent arguments via stdin")
	flagSet.StringVar(&pathOverride, "path", "", "Providing the path argument overrides the $PWD of the pipeline for generation")
	flagSet.StringVar(&version, "version", "latest", "The version is provided by the 'scribe' command, however if only using 'go run', it can be provided here")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		return nil, err
	}

	// Validation: `-step` is mutually exclusive with `-p`.
	if step.Value != 0 && len(pipelineName.names) != 0 {
		return nil, errors.New("both '-step' and '-pipeline' (-p) can not be provided at the same time")
	}

	arguments := &PipelineArgs{
		CanStdinPrompt: !noStdinPrompt,
		Client:         client,
		Version:        version,
		LogLevel:       level,
		BuildID:        buildID,
		State:          state,
		PipelineName:   pipelineName.names,
		Event:          event,
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
