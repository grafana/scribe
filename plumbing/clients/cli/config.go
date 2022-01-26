package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

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

func currentCommit() (string, error) {
	v, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(v))
	}

	return string(v), nil
}

func currentBranch() (string, error) {
	v, err := exec.Command("git", "rev-parse", "--abrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w. output: %s", err, string(v))
	}

	return string(v), nil
}

var KnownValues = map[types.StepArgument]func() (string, error){
	types.ArgumentRemoteURL: currentRemote,
	types.ArgumentCommitRef: currentCommit,
	types.ArgumentBranch:    currentBranch,
}

// Retrieving a config value when using the CLI client will look for arguments to be provided in the `-arg={key}={value}`.
// If they are not available there, then the pipeline will prompt for the value of the argument by requesting input via stdin.
// If the argument "-no-stdin" is provided, then an error will returned instead.
// Some arguments can be assumed by the current environment. When running in CLI mode, for example, the following is almost always true, like:
// * You're _probably_ in the git repo and on the commit that you want to test already
func (c *Client) Value(arg types.StepArgument) (string, error) {
	args := c.Opts.Args.ArgMap

	if val, ok := args[string(arg)]; ok {
		return val, nil
	}

	if argFunc, ok := KnownValues[arg]; ok {
		val, err := argFunc()
		if err == nil {
			return val, nil
		}

		plog.Warnf("shipwright attempted to automatically populate the argument '%s', but encountered an error '%s'", arg, err)
	}

	errMissingArgument := fmt.Errorf("%w: Requested argument '%s'", plumbing.ErrorMissingArgument, string(arg))
	if !c.Opts.Args.CanStdinPrompt {
		return "", errMissingArgument
	}

	fmt.Fprintf(os.Stdout, "Argument '%[1]s' requested but not found. Please provide a value for '%[1]s': ", string(arg))
	// Prompt for the value via stdin since it was not found
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", err
	}

	value := scanner.Text()
	fmt.Fprintf(os.Stdout, "In the future, you can provide this value with the '-arg=%s=' argument\n", value)
	return value, nil
}
