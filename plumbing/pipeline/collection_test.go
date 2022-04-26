package pipeline_test

import (
	"context"
	"testing"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/dag"
	"github.com/grafana/shipwright/plumbing/testutil"
)

func TestCollectionAddSteps(t *testing.T) {
	t.Run("AddSteps should add steps to the graph", func(t *testing.T) {
		col := shipwright.NewDefaultCollection(pipeline.CommonOpts{
			Name: "test",
		})
		steps := pipeline.Step[pipeline.StepList]{
			Serial: 3,
			Content: []pipeline.Step[pipeline.Action]{
				{
					Serial: 1,
					Name:   "step 1",
				},
				{
					Serial: 2,
					Name:   "step 2",
				},
			},
		}

		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, steps), nil)
	})

	t.Run("AddSteps should add steps to the graph with the correct edges", func(t *testing.T) {
		col := shipwright.NewDefaultCollection(pipeline.CommonOpts{
			Name: "test",
		})
		step1 := pipeline.Step[pipeline.StepList]{
			Serial: 7,
			Content: []pipeline.Step[pipeline.Action]{
				{
					Serial: 1,
					Name:   "step 1",
				},
				{
					Serial: 2,
					Name:   "step 2",
				},
			},
		}

		step2 := pipeline.Step[pipeline.StepList]{
			Serial:       8,
			Dependencies: []pipeline.Step[pipeline.StepList]{step1},
			Content: []pipeline.Step[pipeline.Action]{
				{
					Serial: 3,
					Name:   "step 3",
				},
				{
					Serial: 4,
					Name:   "step 4",
				},
				{
					Serial: 5,
					Name:   "step 5",
				},
			},
		}

		step3 := pipeline.Step[pipeline.StepList]{
			Serial:       9,
			Dependencies: []pipeline.Step[pipeline.StepList]{step2},
			Content: pipeline.StepList{
				{
					Serial: 6,
					Name:   "step 6",
				},
			},
		}

		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step1), nil)
		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step2), nil)
		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step3), nil)

		expectedEdges := map[int64][]int64{
			0: {7},
			7: {8},
			8: {9},
		}

		g, _ := col.Graph.Node(shipwright.DefaultPipelineID)
		dag.EnsureGraphEdges(t, expectedEdges, g.Value.Content.Edges)
	})

	t.Run("AddSteps should always make steps where type == StepTypeBackground a child of the root node", func(t *testing.T) {
		col := shipwright.NewDefaultCollection(pipeline.CommonOpts{
			Name: "test",
		})
		step1 := pipeline.Step[pipeline.StepList]{
			Serial: 1,
			Content: pipeline.StepList{
				{
					Serial: 2,
					Name:   "step 1",
				},
				{
					Serial: 3,
					Name:   "step 2",
				},
			},
		}

		step2 := pipeline.Step[pipeline.StepList]{
			Serial: 4,
			Content: pipeline.StepList{
				{
					Serial: 5,
					Name:   "step 3",
					Type:   pipeline.StepTypeBackground,
				},
				{
					Serial: 6,
					Name:   "step 4",
					Type:   pipeline.StepTypeBackground,
				},
				{
					Serial: 7,
					Name:   "step 5",
					Type:   pipeline.StepTypeBackground,
				},
			},
		}

		step3 := pipeline.Step[pipeline.StepList]{
			Serial:       8,
			Dependencies: []pipeline.Step[pipeline.StepList]{step1},
			Content: pipeline.StepList{
				{
					Serial: 9,
					Name:   "step 6",
				},
			},
		}

		// Add 1, 2
		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step1), nil)

		// Add 3, 4, 5
		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step2), nil)

		// Add 6
		testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step3), nil)

		expectedEdges := map[int64][]int64{
			0: {1, 4},
			1: {8},
		}

		g, _ := col.Graph.Node(shipwright.DefaultPipelineID)

		dag.EnsureGraphEdges(t, expectedEdges, g.Value.Content.Edges)
	})
}

func TestCollectionGetters(t *testing.T) {
	col := shipwright.NewDefaultCollection(pipeline.CommonOpts{
		Name: "test",
	})
	step1 := pipeline.Step[pipeline.StepList]{
		Serial: 1,
		Content: pipeline.StepList{
			{
				Serial: 2,
				Name:   "step 1",
			},
			{
				Serial: 3,
				Name:   "step 2",
			},
		},
	}

	step2 := pipeline.Step[pipeline.StepList]{
		Serial: 4,
		Content: pipeline.StepList{
			{
				Serial: 5,
				Name:   "step 3",
				Type:   pipeline.StepTypeBackground,
			},
			{
				Serial: 6,
				Name:   "step 4",
				Type:   pipeline.StepTypeBackground,
			},
			{
				Serial: 7,
				Name:   "step 5",
				Type:   pipeline.StepTypeBackground,
			},
		},
	}

	step3 := pipeline.Step[pipeline.StepList]{
		Serial:       8,
		Dependencies: []pipeline.Step[pipeline.StepList]{step1},
		Content: pipeline.StepList{
			{
				Serial: 9,
				Name:   "step 6",
			},
		},
	}

	// Add 1, 2
	testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step1), nil)

	// Add 3, 4, 5
	testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step2), nil)

	// Add 6
	testutil.EnsureError(t, col.AddSteps(shipwright.DefaultPipelineID, step3), nil)

	t.Run("BySerial should return the step that has the provided serial number", func(t *testing.T) {
		steps, err := col.BySerial(context.Background(), 9)
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

func TestCollectionSub(t *testing.T) {
}
