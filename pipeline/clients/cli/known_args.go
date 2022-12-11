package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/stringutil"
)

// This function effectively runs 'git remote get-url $(git remote)'
func setCurrentRemote(s state.Writer) error {
	remote, err := exec.Command("git", "remote").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(remote))
	}

	v, err := exec.Command("git", "remote", "get-url", strings.TrimSpace(string(remote))).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(pipeline.ArgumentRemoteURL, string(v))
}

// This function effectively runs 'git rev-parse HEAD'
func setCurrentCommit(s state.Writer) error {
	v, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(pipeline.ArgumentCommitRef, string(v))
}

// This function effectively runs 'git rev-parse --abrev-ref HEAD'
func setCurrentBranch(s state.Writer) error {
	v, err := exec.Command("git", "rev-parse", "--abrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w. output: %s", err, string(v))
	}

	return s.SetString(pipeline.ArgumentBranch, string(v))
}

func setWorkingDir(s state.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return s.SetString(pipeline.ArgumentWorkingDir, wd)
}

func setSourceFS(s state.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	return s.SetDirectory(pipeline.ArgumentSourceFS, wd)
}

func setBuildID(s state.Writer) error {
	r := stringutil.Random(8)
	return s.SetString(pipeline.ArgumentBuildID, r)
}

// KnownValues are URL values that we know how to retrieve using the command line.
var KnownValues = map[state.Argument]func(state.Writer) error{
	pipeline.ArgumentRemoteURL:  setCurrentRemote,
	pipeline.ArgumentCommitRef:  setCurrentCommit,
	pipeline.ArgumentBranch:     setCurrentBranch,
	pipeline.ArgumentWorkingDir: setWorkingDir,
	pipeline.ArgumentSourceFS:   setSourceFS,
	pipeline.ArgumentBuildID:    setBuildID,
}
