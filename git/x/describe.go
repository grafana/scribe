package x

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/grafana/shipwright/exec"
)

func Describe(ctx context.Context, dir string, tags bool, dirty bool, always bool) (string, error) {
	var (
		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)
	)

	args := []string{"describe"}
	if tags {
		args = append(args, "--tags")
	}
	if dirty {
		args = append(args, "--dirty")
	}
	if always {
		args = append(args, "--always")
	}

	if err := exec.RunCommandAt(ctx, stdout, stderr, dir, "git", args...); err != nil {
		return "", fmt.Errorf("%w\n%s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
