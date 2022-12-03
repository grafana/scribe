package commands

import "github.com/grafana/scribe/args"

// MustParseRunArgs parses the "run" arguments from the args slice. These options are provided by the scribe command and are typically not user-specified
func MustParseArgs(pargs []string) *args.PipelineArgs {
	v, err := args.ParseArguments(pargs)
	if err != nil {
		panic(err)
	}

	return v
}
