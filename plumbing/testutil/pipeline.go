package testutil

import (
	"context"
	"io"
	"testing"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
)

func RunPipeline(ctx context.Context, t *testing.T, stdout io.Writer, stderr io.Writer, args *plumbing.Arguments) {
	if err := commands.Run(ctx, stdout, stderr, args); err != nil {
		t.Fatal(err)
	}
}
