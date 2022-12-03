package pipeline

import "github.com/grafana/scribe/state"

// These arguments are the pre-defined ones and are mostly used in events.
var (

	// Git arguments
	ArgumentCommitSHA = state.NewStringArgument("git-commit-sha")
	ArgumentCommitRef = state.NewStringArgument("git-commit-ref")
	ArgumentBranch    = state.NewStringArgument("git-branch")
	ArgumentRemoteURL = state.NewStringArgument("remote-url")
	ArgumentTagName   = state.NewStringArgument("git-tag")

	// Standard pipeline arguments
	ArgumentWorkingDir = state.NewStringArgument("workdir")
	// ArgumentSourceFS is the path to the root of the source code for this project.
	ArgumentSourceFS        = state.NewUnpackagedDirectoryArgument("source")
	ArgumentPipelineGoModFS = state.NewUnpackagedDirectoryArgument("pipeline-go-mod")
	ArgumentDockerSocketFS  = state.NewUnpackagedDirectoryArgument("docker-socket")

	// CI service arguments
	ArgumentBuildID = state.NewStringArgument("build-id")
)
