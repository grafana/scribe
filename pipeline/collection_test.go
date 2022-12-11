package pipeline_test

import (
	"context"
	"testing"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/testutil"
)

func TestCollectionAddEvents(t *testing.T) {
	t.Run("AddEvents should add an event to the pipeline", func(t *testing.T) {
		col := scribe.NewDefaultCollection(clients.CommonOpts{
			Name: "test",
		})
		events := []pipeline.Event{
			pipeline.GitCommitEvent(pipeline.GitCommitFilters{}),
			pipeline.GitTagEvent(pipeline.GitTagFilters{}),
		}

		testutil.EnsureError(t, col.AddEvents(scribe.DefaultPipelineID, events...), nil)

		node, err := col.Graph.Node(scribe.DefaultPipelineID)
		testutil.EnsureError(t, err, nil)
		if len(node.Value.Events) != len(events) {
			t.Fatalf("Unexpected number of events in pipeline. Expected '%d', found '%d", len(events), len(node.Value.Events))
		}
	})
	t.Run("Walking a pipeline should have the pipeline events", func(t *testing.T) {
		col := scribe.NewDefaultCollection(clients.CommonOpts{
			Name: "test",
		})
		events := []pipeline.Event{
			pipeline.GitCommitEvent(pipeline.GitCommitFilters{}),
			pipeline.GitTagEvent(pipeline.GitTagFilters{}),
		}

		testutil.EnsureError(t, col.AddEvents(scribe.DefaultPipelineID, events...), nil)

		col.WalkPipelines(context.Background(), func(ctx context.Context, p ...pipeline.Pipeline) error {
			for _, v := range p {
				if len(v.Events) != len(events) {
					t.Fatalf("Expected '%d' events but found '%d' in pipeline", len(events), len(v.Events))
				}
			}

			return nil
		})
	})
}

func TestCollectionAddPipeline(t *testing.T) {
}

func TestCollectionAddSteps(t *testing.T) {
	t.Run("AddSteps should add steps to the graph", func(t *testing.T) {
		col := scribe.NewDefaultCollection(clients.CommonOpts{
			Name: "test",
		})
		steps := []pipeline.Step{
			{
				ID:   1,
				Name: "step 1",
			},
			{
				ID:   2,
				Name: "step 2",
			},
		}

		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, steps...), nil)
	})

	t.Run("AddSteps should add steps to the graph with the correct edges", func(t *testing.T) {
		col := scribe.NewDefaultCollection(clients.CommonOpts{
			Name: "test",
		})
		step1 := []pipeline.Step{
			{
				ID:   1,
				Name: "step 1",
			},
			{
				ID:   2,
				Name: "step 2",
			},
		}

		step2 := []pipeline.Step{
			{
				ID:   3,
				Name: "step 3",
			},
			{
				ID:   4,
				Name: "step 4",
			},
			{
				ID:   5,
				Name: "step 5",
			},
		}

		step3 := []pipeline.Step{
			{
				ID:   6,
				Name: "step 6",
			},
		}

		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step1...), nil)
		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step2...), nil)
		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step3...), nil)
	})

	t.Run("AddSteps should always make steps where type == StepTypeBackground a child of the root node", func(t *testing.T) {
		col := scribe.NewDefaultCollection(clients.CommonOpts{
			Name: "test",
		})
		step1 := []pipeline.Step{
			{
				ID:   2,
				Name: "step 1",
			},
			{
				ID:   3,
				Name: "step 2",
			},
		}

		step2 := []pipeline.Step{
			{
				ID:   5,
				Name: "step 3",
				Type: pipeline.StepTypeBackground,
			},
			{
				ID:   6,
				Name: "step 4",
				Type: pipeline.StepTypeBackground,
			},
			{
				ID:   7,
				Name: "step 5",
				Type: pipeline.StepTypeBackground,
			},
		}

		step3 := []pipeline.Step{
			{
				ID:   9,
				Name: "step 6",
			},
		}

		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step1...), nil)
		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step2...), nil)
		testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step3...), nil)
	})
}

func TestCollectionGetters(t *testing.T) {
	col := scribe.NewDefaultCollection(clients.CommonOpts{
		Name: "test",
	})

	step1 := []pipeline.Step{
		{
			ID:   2,
			Name: "step 1",
		},
		{
			ID:   3,
			Name: "step 2",
		},
	}

	step2 := []pipeline.Step{
		{
			ID:   5,
			Name: "step 3",
			Type: pipeline.StepTypeBackground,
		},
		{
			ID:   6,
			Name: "step 4",
			Type: pipeline.StepTypeBackground,
		},
		{
			ID:   7,
			Name: "step 5",
			Type: pipeline.StepTypeBackground,
		},
	}

	step3 := []pipeline.Step{
		{
			ID:   9,
			Name: "step 6",
		},
	}

	testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step1...), nil)
	testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step2...), nil)
	testutil.EnsureError(t, col.AddSteps(scribe.DefaultPipelineID, step3...), nil)

	t.Run("ByID should return the step that has the provided ID", func(t *testing.T) {
		steps, err := col.ByID(context.Background(), 9)
		if err != nil {
			t.Fatal(err)
		}

		if len(steps) != 1 {
			t.Fatalf("expected 1 step but got '%d'", len(steps))
		}

		if steps[0].Name != "step 6" {
			t.Fatalf("expected step to be 'step 6', but got '%v'", steps[0])
		}
	})

	t.Run("ByName should return the step that has the provided name", func(t *testing.T) {
		steps, err := col.ByName(context.Background(), "step 6")
		if err != nil {
			t.Fatal(err)
		}

		if len(steps) != 1 {
			t.Fatalf("expected 1 step but got '%d'", len(steps))
		}

		if steps[0].Name != "step 6" {
			t.Fatalf("expected step to be 'step 6', but got '%v'", steps[0])
		}
	})
}

func TestCollectionByName(t *testing.T) {
}
