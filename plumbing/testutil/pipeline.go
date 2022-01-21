package testutil

import (
	"bytes"
	"context"
	"io"
	"testing"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

func RunPipeline(ctx context.Context, t *testing.T, path string, stdout io.Writer, stderr io.Writer, args *plumbing.Arguments) {
	buf := bytes.NewBuffer(nil)
	t.Log("Running pipeline with args", args)
	if err := commands.Run(ctx, path, stdout, io.MultiWriter(stderr, buf), args); err != nil {
		t.Fatalf("Error running pipeline. Error: '%s'\nStderr: '%s'\n", err, buf.String())
	}
}

// NewTestStep creates a new TestStep that emits data into the channel 'b' when the action is ran
func NewTestStep(b chan bool) types.Step {
	return types.Step{
		Name: "test",
		Action: func() error {
			b <- true
			return nil
		},
	}
}
