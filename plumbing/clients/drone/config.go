package drone

import (
	"fmt"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

var argEnvMap = map[types.StepArgument]string{
	types.ArgumentCommitSHA:  "$DRONE_COMMIT",
	types.ArgumentCommitRef:  "$DRONE_COMMIT_REF",
	types.ArgumentRemoteURL:  "$DRONE_GIT_SSH_URL",
	types.ArgumentWorkingDir: "$DRONE_REPO_NAME",
}

func (c *Client) Value(arg types.StepArgument) (string, error) {
	if val, ok := argEnvMap[arg]; ok {
		return val, nil
	}

	return "", fmt.Errorf("could not find equivalent of '%s': %w", string(arg), plumbing.ErrorMissingArgument)
}
