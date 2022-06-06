package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/scribe/plumbing/pipeline"
)

// This function effectively runs 'git remote get-url $(git remote)'
func setCurrentRemote(state *pipeline.State) error {
	remote, err := exec.Command("git", "remote").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(remote))
	}

	v, err := exec.Command("git", "remote", "get-url", strings.TrimSpace(string(remote))).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return state.SetString(pipeline.ArgumentRemoteURL, string(v))
}

// This function effectively runs 'git rev-parse HEAD'
func setCurrentCommit(state *pipeline.State) error {
	v, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return state.SetString(pipeline.ArgumentCommitRef, string(v))
}

// This function effectively runs 'git rev-parse --abrev-ref HEAD'
func setCurrentBranch(state *pipeline.State) error {
	v, err := exec.Command("git", "rev-parse", "--abrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return state.SetString(pipeline.ArgumentBranch, string(v))
}

func setWorkingDir(state *pipeline.State) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return state.SetString(pipeline.ArgumentWorkingDir, wd)
}

func setSourceFS(state *pipeline.State) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return state.SetDirectory(pipeline.ArgumentSourceFS, wd)
}

// KnownValues are URL values that we know how to retrieve using the command line.
var KnownValues = map[pipeline.Argument]func(*pipeline.State) error{
	pipeline.ArgumentRemoteURL:  setCurrentRemote,
	pipeline.ArgumentCommitRef:  setCurrentCommit,
	pipeline.ArgumentBranch:     setCurrentBranch,
	pipeline.ArgumentWorkingDir: setWorkingDir,
	pipeline.ArgumentSourceFS:   setSourceFS,
}

// // GetArgValue first checks for the arguments provided in the `-arg` CLI argument. If one is found, then it returns that.
// // If one is not found, then it checks the KnownValues map above to try to determine the value without asking for it.
// // This is likely to be deprecated in the future whenever these arguments can be populated from events.
// func GetArgValue(args *plumbing.PipelineArgs, arg pipeline.Argument) (string, error) {
// 	if val, ok := args.ArgMap[arg.Key]; ok {
// 		return val, nil
// 	}
//
// 	if argFunc, ok := KnownValues[arg]; ok {
// 		val, err := argFunc()
// 		if err == nil {
// 			return val, nil
// 		}
// 	}
//
// 	return "", fmt.Errorf("%w: Requested argument '%s'", plumbing.ErrorMissingArgument, arg.Key)
// 	// errMissingArgument := fmt.Errorf("%w: Requested argument '%s'", plumbing.ErrorMissingArgument, arg.Key)
// 	// if !args.CanStdinPrompt {
// 	// 	return "", errMissingArgument
// 	// }
//
// 	// fmt.Fprintf(os.Stdout, "Argument '%[1]s' requested but not found. Please provide a value for '%[1]s': ", arg.Key)
// 	// // Prompt for the value via stdin since it was not found
// 	// scanner := bufio.NewScanner(os.Stdin)
// 	// scanner.Scan()
//
// 	// if err := scanner.Err(); err != nil {
// 	// 	return "", err
// 	// }
//
// 	// value := scanner.Text()
// 	// fmt.Fprintf(os.Stdout, "In the future, you can provide this value with the '-arg=%s=%s' argument\n", arg.Key, value)
// 	// return value, nil
// }
//
// // Retrieving a config value when using the CLI client will look for arguments to be provided in the `-arg={key}={value}`.
// // If they are not available there, then the pipeline will prompt for the value of the argument by requesting input via stdin.
// // If the argument "-no-stdin" is provided, then an error will returned instead.
// // Some arguments can be assumed by the current environment. When running in CLI mode, for example, the following is almost always true, like:
// // * You're _probably_ in the git repo and on the commit that you want to test already
// func (c *Client) Value(arg pipeline.Argument) (string, error) {
// 	return GetArgValue(c.Opts.Args, arg)
// }
