package pipeline

// These arguments are the pre-defined ones and are mostly used in events.
var (
	ArgumentDockerSocketFS = NewUnpackagedDirectoryArgument("docker-socket")

	// Git arguments
	ArgumentCommitSHA = NewStringArgument("git-commit-sha")
	ArgumentCommitRef = NewStringArgument("git-commit-ref")
	ArgumentBranch    = NewStringArgument("git-branch")
	ArgumentRemoteURL = NewStringArgument("remote-url")
	ArgumentTagName   = NewStringArgument("git-tag")

	// Standard pipeline arguments
	ArgumentWorkingDir = NewStringArgument("workdir")
	// ArgumentSourceFS is the path to the root of the source code for this project.
	ArgumentSourceFS = NewUnpackagedDirectoryArgument("source")

	// CI service arguments
	ArgumentBuildID = NewStringArgument("build-id")
)
