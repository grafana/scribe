package types

// A StepArgument is a pre-defined argument that is used in a typical CI pipeline.
// This allows the shipwright library to define different methods of retrieving the same information
// in different run modes.
// For example, when running in CLI or Docker mode, getting the git ref might be as simple as running `git rev-parse HEAD`.
// But in a Drone pipeline, that information may be available before the repository has been cloned in an environment variable.
// Other arguments may require the user to be prompted if they have not been provided.
// These arguments can be provided to the CLI by using the flag `-arg`, for example `-arg=workdir=./example` will set the "workdir" argument to "example" in the CLI run-mode
type StepArgument string

// These arguments are the pre-defined ones.
const (
	ArgumentCommitSHA  = "git-commit-sha"
	ArgumentCommitRef  = "git-commit-ref"
	ArgumentBranch     = "git-branch"
	ArgumentWorkingDir = "workdir"
	ArgumentRemoteURL  = "remote-url"
)
