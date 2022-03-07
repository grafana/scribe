package drone

import (
	"fmt"

	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

var argEnvMap = map[pipeline.Argument]string{
	pipeline.ArgumentCommitSHA:  "$DRONE_COMMIT",
	pipeline.ArgumentCommitRef:  "$DRONE_COMMIT_REF",
	pipeline.ArgumentRemoteURL:  "$DRONE_GIT_SSH_URL",
	pipeline.ArgumentWorkingDir: "$DRONE_REPO_NAME",
}

func (c *Client) Value(arg pipeline.Argument) (string, error) {
	if val, ok := argEnvMap[arg]; ok {
		return val, nil
	}

	return "", fmt.Errorf("could not find equivalent of '%s': %w", arg.Key, plumbing.ErrorMissingArgument)
}
