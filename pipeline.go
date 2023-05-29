package scribe

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

type ActionOpts struct {
	Logger slog.Logger
	State  StateHandler
	Stdout io.Writer
	Stderr io.Writer
}

// StepActionFunc is the action that is executed when the arguments from this step are needed.
type PipelineActionFunc func(ctx context.Context, opts ActionOpts) error

// A Pipeline is an entire dagger pipeline.
// A Pipeline has:
// * an action, which is the dagger pipeline that is executed for this step
// * requirements, which are things that must exist in the state before this pipeline can be executed
// * outputs, which are things that this pipeline provides into the state to satisfy the requirements of other pipelines
type Pipeline struct {
	Name     string
	Action   PipelineActionFunc
	Requires []Argument
	Provides []Argument
}
