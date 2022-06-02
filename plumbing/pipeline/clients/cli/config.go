package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
)

// This function effectively runs 'git remote get-url $(git remote)'
func currentRemote() (string, error) {
	remote, err := exec.Command("git", "remote").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(remote))
	}

	v, err := exec.Command("git", "remote", "get-url", strings.TrimSpace(string(remote))).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(v))
	}

	return string(v), nil
}

// This function effectively runs 'git rev-parse HEAD'
func currentCommit() (string, error) {
	v, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(v))
	}

	return string(v), nil
}

// This function effectively runs 'git rev-parse --abrev-ref HEAD'
func currentBranch() (string, error) {
	v, err := exec.Command("git", "rev-parse", "--abrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(v))
	}

	return string(v), nil
}

// KnownValues are URL values that we know how to retrieve using the command line.
var KnownValues = map[pipeline.Argument]func() (string, error){
	pipeline.ArgumentRemoteURL:  currentRemote,
	pipeline.ArgumentCommitRef:  currentCommit,
	pipeline.ArgumentBranch:     currentBranch,
	pipeline.ArgumentWorkingDir: os.Getwd,
}

func GetArgValue(args *plumbing.PipelineArgs, arg pipeline.Argument) (string, error) {
	if val, ok := args.ArgMap[arg.Key]; ok {
		return val, nil
	}

	if argFunc, ok := KnownValues[arg]; ok {
		val, err := argFunc()
		if err == nil {
			return val, nil
		}
	}

	return "", fmt.Errorf("%w: Requested argument '%s'", plumbing.ErrorMissingArgument, arg.Key)
	// errMissingArgument := fmt.Errorf("%w: Requested argument '%s'", plumbing.ErrorMissingArgument, arg.Key)
	// if !args.CanStdinPrompt {
	// 	return "", errMissingArgument
	// }

	// fmt.Fprintf(os.Stdout, "Argument '%[1]s' requested but not found. Please provide a value for '%[1]s': ", arg.Key)
	// // Prompt for the value via stdin since it was not found
	// scanner := bufio.NewScanner(os.Stdin)
	// scanner.Scan()

	// if err := scanner.Err(); err != nil {
	// 	return "", err
	// }

	// value := scanner.Text()
	// fmt.Fprintf(os.Stdout, "In the future, you can provide this value with the '-arg=%s=%s' argument\n", arg.Key, value)
	// return value, nil
}

// Retrieving a config value when using the CLI client will look for arguments to be provided in the `-arg={key}={value}`.
// If they are not available there, then the pipeline will prompt for the value of the argument by requesting input via stdin.
// If the argument "-no-stdin" is provided, then an error will returned instead.
// Some arguments can be assumed by the current environment. When running in CLI mode, for example, the following is almost always true, like:
// * You're _probably_ in the git repo and on the commit that you want to test already
func (c *Client) Value(arg pipeline.Argument) (string, error) {
	return GetArgValue(c.Opts.Args, arg)
}
