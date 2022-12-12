package scribe_test

import (
	"context"
	"testing"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

func TestMulti(t *testing.T) {
	t.Run("Multi pipelines should have a root node with an ID of zero", func(t *testing.T) {
		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := scribe.NewMultiWithClient(testOpts, newEnsurer())

		if sw.Collection == nil {
			t.Fatal("Collection is nil")
		}

		_, err := sw.Collection.Graph.Node(0)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Creating a multi-pipeline with steps", func(t *testing.T) {
		// This is a potentially flaky test because I think 4 could sometimes show up before 5.
		ens := newEnsurer(
			"step 1", "step 2", "step 3", "step 5", "step 4",
			"step 1", "step 2", "step 3", "step 5", "step 4",
		)

		// In this test case we're not providing ensurer data because we are not running 'Done'.
		var (
			sw   = scribe.NewMultiWithClient(testOpts, ens)
			argA = state.NewStringArgument("a")
			argB = state.NewStringArgument("b")
			argC = state.NewStringArgument("c")
		)
		mf := func(s *scribe.Scribe) {
			s.Add(
				pipeline.NoOpStep.WithName("step 1").Provides(argA),
				pipeline.NoOpStep.WithName("step 2").Provides(argC),
			)
			s.Add(
				pipeline.NoOpStep.WithName("step 3").Requires(argA).Provides(argB),
			)
			s.Add(
				pipeline.NoOpStep.WithName("step 4").Requires(argB, argC),
				pipeline.NoOpStep.WithName("step 5").Requires(argA, argB, argC),
			)
		}

		// each multi-func adds 5 new steps, and each new sub-pipeline adds an additional root step.
		// These pipelines are processed after all of the others are, so they will have the highest IDs (23 and 24).
		sw.Add(
			sw.New("test 1", mf),
			sw.New("test 2", mf),
		)

		if err := sw.Execute(context.Background(), sw.Collection); err != nil {
			t.Fatal(err)
		}
	})
}

func TestMultiWithEvent(t *testing.T) {
	t.Run("Once adding an event, it should be present in the collection", func(t *testing.T) {
		ens := newEnsurer()
		sw := scribe.NewMultiWithClient(testOpts, ens)

		mf := func(sw *scribe.Scribe) {
			sw.When(
				pipeline.GitTagEvent(pipeline.GitTagFilters{}),
			)

			sw.Add(pipeline.NoOpStep.WithName("step 1"))
		}

		sw.Add(
			sw.New("test 1", mf),
		)

		sw.Collection.WalkPipelines(context.Background(), func(ctx context.Context, pipelines ...pipeline.Pipeline) error {
			for _, v := range pipelines {
				if len(v.Events) != 1 {
					t.Fatal("Expected 1 pipeline event, but found", len(v.Events))
				}
			}

			return nil
		})
	})
}
