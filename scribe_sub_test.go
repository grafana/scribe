package scribe_test

import (
	"context"
	"testing"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/plumbing/pipeline"
)

func TestSub(t *testing.T) {
	t.Run("A standard sub-pipeline should render steps in the appropriate order", func(t *testing.T) {
		ens := newEnsurer(
			[]string{"step 1"}, []string{"step 2"},
			[]string{"step 6"}, []string{"step 7"},
			[]string{"step 3"},
			[]string{"step 4"}, []string{"step 5"},
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := scribe.NewWithClient(testOpts, ens)

		sw.Run(pipeline.NoOpStep.WithName("step 1"), pipeline.NoOpStep.WithName("step 2"))

		sf := func(sw *scribe.Scribe) {
			sw.Run(pipeline.NoOpStep.WithName("step 3"))
			sw.Run(pipeline.NoOpStep.WithName("step 4"), pipeline.NoOpStep.WithName("step 5"))
		}

		sw.Sub(sf)

		sw.Run(pipeline.NoOpStep.WithName("step 6"), pipeline.NoOpStep.WithName("step 7"))

		if err := sw.Execute(context.Background(), sw.Collection); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("A standard sub-pipeline should render parallel steps in the appropriate order", func(t *testing.T) {
		ens := newEnsurer(
			[]string{"step 1"}, []string{"step 2"},
			[]string{"step 6"}, []string{"step 7"},
			[]string{"step 3", "step 4"},
			[]string{"step 5"},
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := scribe.NewWithClient(testOpts, ens)

		sw.Run(pipeline.NoOpStep.WithName("step 1"), pipeline.NoOpStep.WithName("step 2"))

		sf := func(sw *scribe.Scribe) {
			sw.Parallel(pipeline.NoOpStep.WithName("step 3"), pipeline.NoOpStep.WithName("step 4"))
			sw.Run(pipeline.NoOpStep.WithName("step 5"))
		}

		sw.Sub(sf)

		sw.Run(pipeline.NoOpStep.WithName("step 6"), pipeline.NoOpStep.WithName("step 7"))

		if err := sw.Execute(context.Background(), sw.Collection); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("A multi-pipeline with a sub-pipeline should render steps in the appropriate order", func(t *testing.T) {
		ens := newEnsurer(
			[]string{"step 1"}, []string{"step 2"}, // test 1
			[]string{"step 1"}, []string{"step 2"}, // test 2
			[]string{"step 1"}, []string{"step 2"}, // test 3
			[]string{"step 1"}, []string{"step 2"}, // test 4
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := scribe.NewMultiWithClient(testOpts, ens)

		mf := func(sw *scribe.Scribe) {
			sw.Run(pipeline.NoOpStep.WithName("step 1"), pipeline.NoOpStep.WithName("step 2"))
		}

		// each multi-func adds 5 new steps, and each new sub-pipeline adds an additional root step.
		// These pipelines are processed after all of the others are, so they will have the highest IDs (23 and 24).
		sw.Run(
			sw.New("test 1", mf),
		)

		sw.Sub(func(sw *scribe.ScribeMulti) {
			sw.Run(
				sw.New("test 2", mf),
				sw.New("test 3", mf),
				sw.New("test 4", mf),
			)
		})

		if err := sw.Execute(context.Background(), sw.Collection); err != nil {
			t.Fatal(err)
		}
	})
}
