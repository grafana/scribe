package cli

import (
	"bufio"
	"fmt"
	"os"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// Retrieving a config value when using the CLI client will look for arguments to be provided in the `-arg={key}={value}`.
// If they are not available there, then the pipeline will prompt for the value of the argument by requesting input via stdin.
// If the argument "-no-stdin" is provided, then an error will returned instead.
func (c *Client) Value(arg types.StepArgument) (string, error) {
	args := c.Opts.Args.ArgMap

	if val, ok := args[string(arg)]; ok {
		return val, nil
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
