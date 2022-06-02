package pipeline

type ArgumentType int

const (
	ArgumentTypeString ArgumentType = iota
	ArgumentTypeInt64
	ArgumentTypeFloat64
	ArgumentTypeSecret
	ArgumentTypeFile
	ArgumentTypeFS
)

var argumentTypeStr = []string{"string", "int", "float", "secret", "file", "directory"}

func (a ArgumentType) String() string {
	i := int(a)
	return argumentTypeStr[i]
}

func ArgumentTypesEqual(argType ArgumentType, arg Argument) bool {
	return arg.Type == argType
}

// An Argument is a pre-defined argument that is used in a typical CI pipeline.
// This allows the scribe library to define different methods of retrieving the same information
// in different run modes.
// For example, when running in CLI or Docker mode, getting the git ref might be as simple as running `git rev-parse HEAD`.
// But in a Drone pipeline, that information may be available before the repository has been cloned in an environment variable.
// Other arguments may require the user to be prompted if they have not been provided.
// These arguments can be provided to the CLI by using the flag `-arg`, for example `-arg=workdir=./example` will set the "workdir" argument to "example" in the CLI run-mode
// By default, all steps expect a WorkingDir and Repository.
type Argument struct {
	Type ArgumentType
	Key  string
}

func NewStringArgument(key string) Argument {
	return Argument{
		Type: ArgumentTypeString,
		Key:  key,
	}
}

func NewInt64Argument(key string) Argument {
	return Argument{
		Type: ArgumentTypeInt64,
		Key:  key,
	}
}

func NewFloat64Argument(key string) Argument {
	return Argument{
		Type: ArgumentTypeFloat64,
		Key:  key,
	}
}

func NewFileArgument(key string) Argument {
	return Argument{
		Type: ArgumentTypeFile,
		Key:  key,
	}
}

func NewDirectoryArgument(key string) Argument {
	return Argument{
		Type: ArgumentTypeFS,
		Key:  key,
	}
}

func NewSecretArgument(key string) Argument {
	return Argument{
		Type: ArgumentTypeSecret,
		Key:  key,
	}
}

// These arguments are the pre-defined ones and are mostly used in events.
var (
	ArgumentSourceFS       = NewDirectoryArgument("source")
	ArgumentDockerSocketFS = NewDirectoryArgument("docker-socket")

	// Git arguments
	ArgumentCommitSHA = NewStringArgument("git-commit-sha")
	ArgumentCommitRef = NewStringArgument("git-commit-ref")
	ArgumentBranch    = NewStringArgument("git-branch")
	ArgumentRemoteURL = NewStringArgument("remote-url")
	ArgumentTagName   = NewStringArgument("git-tag")

	// Standard pipeline arguments
	ArgumentWorkingDir = NewStringArgument("workdir")
)
