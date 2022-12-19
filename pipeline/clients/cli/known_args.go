package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/stringutil"
)

// This function effectively runs 'git remote get-url $(git remote)'
func setCurrentRemote(ctx context.Context, s state.Writer) error {
	remote, err := exec.Command("git", "remote").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(remote))
	}

	v, err := exec.Command("git", "remote", "get-url", strings.TrimSpace(string(remote))).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(ctx, pipeline.ArgumentRemoteURL, string(v))
}

// This function effectively runs 'git rev-parse HEAD'
func setCurrentCommit(ctx context.Context, s state.Writer) error {
	v, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(ctx, pipeline.ArgumentCommitRef, string(v))
}

// This function effectively runs 'git rev-parse --abrev-ref HEAD'
func setCurrentBranch(ctx context.Context, s state.Writer) error {
	v, err := exec.Command("git", "rev-parse", "--abrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(ctx, pipeline.ArgumentBranch, string(v))
}

func setWorkingDir(ctx context.Context, s state.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return s.SetString(ctx, pipeline.ArgumentWorkingDir, wd)
}

func setSourceFS(ctx context.Context, s state.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return s.SetDirectory(ctx, pipeline.ArgumentSourceFS, wd)
}

func setBuildID(ctx context.Context, s state.Writer) error {
	r := stringutil.Random(8)
	return s.SetString(ctx, pipeline.ArgumentBuildID, r)
}

// KnownValues are URL values that we know how to retrieve using the command line.
var KnownValues = map[state.Argument]func(context.Context, state.Writer) error{
	pipeline.ArgumentRemoteURL:  setCurrentRemote,
	pipeline.ArgumentCommitRef:  setCurrentCommit,
	pipeline.ArgumentBranch:     setCurrentBranch,
	pipeline.ArgumentWorkingDir: setWorkingDir,
	pipeline.ArgumentSourceFS:   setSourceFS,
	pipeline.ArgumentBuildID:    setBuildID,
}
