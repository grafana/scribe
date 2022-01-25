package commands

import "pkg.grafana.com/shipwright/v1/plumbing"

// MustParseRunArgs parses the "run" arguments from the args slice. These options are provided by the shipwright command and are typically not user-specified
func MustParseArgs(args []string) *plumbing.PipelineArgs {
	v, err := plumbing.ParseArguments(args)
	if err != nil {
		panic(err)
	}

	return v
}
