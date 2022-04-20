package shipwright_test

import (
	"context"
	"testing"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

func TestMulti(t *testing.T) {
	t.Run("Multi pipelines should have a root node with an ID of zero", func(t *testing.T) {
		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := shipwright.NewMultiWithClient[pipeline.Pipeline](testOpts, newEnsurer())

		if sw.Collection == nil {
			t.Fatal("Collection is nil")
		}

		_, err := sw.Collection.Graph.Node(0)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Creating a multi-pipeline with steps", func(t *testing.T) {
		ens := newEnsurer(
			[]string{"step 1"}, []string{"step 2"}, []string{"step 3"}, []string{"step 4"}, []string{"step 5"},
			[]string{"step 1"}, []string{"step 2"}, []string{"step 3"}, []string{"step 4"}, []string{"step 5"},
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := shipwright.NewMultiWithClient[pipeline.Pipeline](testOpts, ens)

		mf := func(sw *shipwright.Shipwright[pipeline.Action]) {
			sw.Run(pipeline.NoOpStep.WithName("step 1"), pipeline.NoOpStep.WithName("step 2"))
			sw.Run(pipeline.NoOpStep.WithName("step 3"))
			sw.Run(pipeline.NoOpStep.WithName("step 4"), pipeline.NoOpStep.WithName("step 5"))
		}

		// each multi-func adds 5 new steps, and each new sub-pipeline adds an additional root step.
		// These pipelines are processed after all of the others are, so they will have the highest IDs (23 and 24).
		sw.Run(
			sw.New("test 1", mf),
			sw.New("test 2", mf),
		)

		if err := sw.Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Should run pipelines in parallel if they are added with the Parallel function", func(t *testing.T) {
		ens := newEnsurer(
			[]string{"step 1", "step 2"}, []string{"step 3", "step 4"},
			[]string{"step 1", "step 2"}, []string{"step 3", "step 4"},
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := shipwright.NewMultiWithClient[pipeline.Pipeline](testOpts, ens)

		mf := func(sw *shipwright.Shipwright[pipeline.Action]) {
			sw.Parallel(pipeline.NoOpStep.WithName("step 1"), pipeline.NoOpStep.WithName("step 2"))
			sw.Parallel(pipeline.NoOpStep.WithName("step 3"), pipeline.NoOpStep.WithName("step 4"))
		}

		// each multi-func adds 5 new steps, and each new sub-pipeline adds an additional root step.
		// These pipelines are processed after all of the others are, so they will have the highest IDs (23 and 24).
		sw.Run(
			sw.New("test 1", mf),
			sw.New("test 2", mf),
		)

		if err := sw.Execute(context.Background()); err != nil {
			t.Fatal(err)
		}
	})
}
