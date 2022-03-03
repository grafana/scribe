package testutil

import (
	"bytes"
	"context"
	"io"
	"testing"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func RunPipeline(ctx context.Context, t *testing.T, path string, stdout io.Writer, stderr io.Writer, args *plumbing.PipelineArgs) {
	buf := bytes.NewBuffer(nil)
	t.Log("Running pipeline with args", args)
	if err := commands.Run(ctx, &commands.RunOpts{
		Path:   path,
		Stdout: stdout,
		Stderr: io.MultiWriter(stderr, buf),
		Args:   args,
	}); err != nil {
		t.Fatalf("Error running pipeline. Error: '%s'\nStderr: '%s'\n", err, buf.String())
	}
}

// NewTestStep creates a new TestStep that emits data into the channel 'b' when the action is ran
func NewTestStep(b chan bool) pipeline.Step {
	return pipeline.Step{
		Name: "test",
		Action: func(context.Context, pipeline.ActionOpts) error {
			b <- true
			return nil
		},
	}
}
