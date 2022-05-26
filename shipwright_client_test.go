package shipwright_test

import (
	"context"
	"fmt"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

// ensurer provides a pipeline.StepWalkFunc that ensures that the steps that it receives are ran in the order provided.
type ensurer struct {
	i     int
	seen  [][]string
	steps [][]string
}

func (e *ensurer) WalkPipelines(w pipeline.Walker) func(context.Context, ...pipeline.Pipeline) error {
	return func(ctx context.Context, pipelines ...pipeline.Pipeline) error {
		for _, v := range pipelines {
			if err := w.WalkSteps(ctx, v.ID, e.WalkSteps); err != nil {
				return err
			}
		}
		return nil
	}
}

func (e *ensurer) WalkSteps(ctx context.Context, steps ...pipeline.Step) error {
	s := make([]string, len(steps))

	for i, v := range steps {
		s[i] = v.Name
	}

	// WalkFunc should should never be called more times than there are lists of steps in e.seen
	if e.i >= len(e.seen) {
		return fmt.Errorf("walk executed more times than expected. received steps '%+v'", s)
	}

	expect := e.steps[e.i]
	if len(s) != len(expect) {
		return fmt.Errorf("unequal number of steps at '%d'. expected steps '%+v', got '%+v'", e.i, expect, s)
	}

	for i, v := range expect {
		if s[i] != v {
			return fmt.Errorf("unexpected step at '%d'. expected step '%s'(%+v), got '%s' (%+v)", e.i, v, expect, s[i], s)
		}
	}

	e.seen[e.i] = s
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

	for i, list := range e.steps {
		if len(list) != len(e.seen[i]) {
			return fmt.Errorf("unequal amount of steps seen at '%d'; expected '%d' (%+v), found '%d' (%+v)\n%s", i, len(list), list, len(e.seen[i]), e.seen[i], e.Diff())
		}

		for n, step := range list {
			if e.seen[i][n] != step {
				return fmt.Errorf("step seen at '%d:%d' does not match expected. Expected '%s', found '%s'\n%s", i, n, e.seen[i][n], step, e.Diff())
			}
		}
	}

	return nil
}

func newEnsurer(steps ...[]string) *ensurer {
	return &ensurer{
		steps: steps,
		seen:  make([][]string, len(steps)),
	}
}
