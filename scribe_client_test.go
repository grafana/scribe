package scribe_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/grafana/scribe/pipeline"
)

// ensurer provides a pipeline.StepWalkFunc that ensures that the steps that it receives are ran in the order provided.
type ensurer struct {
	i     int
	seen  []string
	steps []string
}

func (e *ensurer) WalkPipelines(w pipeline.Walker) func(context.Context, pipeline.Pipeline) error {
	return func(ctx context.Context, p pipeline.Pipeline) error {
		if err := w.WalkSteps(ctx, p.ID, e.WalkSteps); err != nil {
			return err
		}
		return nil
	}
}

func (e *ensurer) WalkSteps(ctx context.Context, step pipeline.Step) error {
	expect := e.steps[e.i]

	if !strings.EqualFold(step.Name, expect) {
		return fmt.Errorf("unexpected step at '%d'. expected step '%s', got '%s'", e.i, expect, step.Name)
	}

	e.seen[e.i] = step.Name
	e.i++

	return nil
}

// Validate is ran internally before calling Run or Parallel and allows the client to effectively configure per-step requirements
// For example, Drone steps MUST have an image so the Drone client returns an error in this function when the provided step does not have an image.
// If the error encountered is not critical but should still be logged, then return a plumbing.ErrorSkipValidation.
// The error is checked with `errors.Is` so the error can be wrapped with fmt.Errorf.
func (e *ensurer) Validate(pipeline.Step) error {
	return nil
}

func (e *ensurer) Diff() string {
	return fmt.Sprintf("Seen:     %+v\nExpected: %+v", e.seen, e.steps)
}

// Done must be ran at the end of the pipeline.
// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
func (e *ensurer) Done(ctx context.Context, w pipeline.Walker) error {
	if err := w.WalkPipelines(ctx, e.WalkPipelines(w)); err != nil {
		return err
	}

	if len(e.seen) != len(e.steps) {
		return fmt.Errorf("walked unequal amount of steps. expected '%d', walked '%d'\n%s", len(e.steps), len(e.seen), e.Diff())
	}

	for i, step := range e.steps {
		if e.seen[i] != step {
			return fmt.Errorf("step seen at '%d' does not match expected. Expected '%s', found '%s'\n%s", i, e.seen[i], step, e.Diff())
		}
	}

	return nil
}

func newEnsurer(steps ...string) *ensurer {
	return &ensurer{
		steps: steps,
		seen:  make([]string, len(steps)),
	}
}
