package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
)

var (
	ArgumentVersion = pipeline.NewStringArgument("version")
)

func version() (string, error) {
	// git config --global --add safe.directory * is needed to resolve the restriction introduced by CVE-2022-24765.
	out, err := exec.Command("git", "config", "--global", "--add", "safe.directory", "*").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running command 'git config --global --add safe.directory *' resulted in error '%w'. Output: '%s'", err, string(out))
	}

	version, err := exec.Command("git", "describe", "--tags", "--dirty", "--always").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running command 'git describe --tags --dirty --always' resulted in the error '%w'. Output: '%s'", err, string(version))
	}

	return strings.TrimSpace(string(version)), nil
}

func getVersion(ctx context.Context, opts pipeline.ActionOpts) error {
	v, err := version()
	if err != nil {
		return err
	}

	return opts.State.SetString(ArgumentVersion, v)
}

func StepGetVersion(version string) pipeline.Step {
	return pipeline.NewStep(getVersion).
		WithArguments(
			pipeline.ArgumentSourceFS,
		).
		Provides(ArgumentVersion).
		WithImage(plumbing.SubImage("git", version))
}
