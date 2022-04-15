package shipwright_test

import (
	"testing"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/dag"
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
		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := shipwright.NewMultiWithClient[pipeline.Pipeline](testOpts, newEnsurer())

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

		t.Run("It should have three nodes", func(t *testing.T) {
			dag.EnsureGraphNodes(t, []int64{0, 23, 24}, sw.Collection.Graph.Nodes)
		})

		t.Run("It should have two edges", func(t *testing.T) {
			dag.EnsureGraphEdges(t, map[int64][]int64{
				0:  {23},
				23: {24},
			}, sw.Collection.Graph.Edges)
		})

		t.Run("The first node should be a graph with 6 nodes", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(23)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphNodes(t, []int64{0, 3, 4, 6, 9, 10}, sub.Value.Content.Nodes)
		})
		t.Run("The first node should be a graph with 5 edges", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(23)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphEdges(t, map[int64][]int64{
				0: {3},
				3: {4},
				4: {6},
				6: {9},
				9: {10},
			}, sub.Value.Content.Edges)
		})

		t.Run("The second node should be a graph with 6 nodes", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(24)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphNodes(t, []int64{0, 14, 15, 17, 20, 21}, sub.Value.Content.Nodes)
		})
		t.Run("The second node should be a graph with 5 edges", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(24)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphEdges(t, map[int64][]int64{
				0:  {14},
				14: {15},
				15: {17},
				17: {20},
				20: {21},
			}, sub.Value.Content.Edges)
		})
	})

	t.Run("Should run pipelines in parallel if they are added with the Parallel function", func(t *testing.T) {
		// In this test case we're not providing ensurer data because we are not running 'Done'.
		sw := shipwright.NewMultiWithClient[pipeline.Pipeline](testOpts, newEnsurer())

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

		t.Run("It should have 3 nodes", func(t *testing.T) {
			dag.EnsureGraphNodes(t, []int64{0, 15, 16}, sw.Collection.Graph.Nodes)
		})

		t.Run("It should have 2 edges", func(t *testing.T) {
			dag.EnsureGraphEdges(t, map[int64][]int64{
				0:  {15},
				15: {16},
			}, sw.Collection.Graph.Edges)
		})

		t.Run("The first node should be a graph with 3 nodes", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(15)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphNodes(t, []int64{0, 3, 6}, sub.Value.Content.Nodes)
		})
		t.Run("The first node should be a graph with 2 edges", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(15)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphEdges(t, map[int64][]int64{
				0: {3},
				3: {6},
			}, sub.Value.Content.Edges)
		})

		t.Run("The second node should be a graph with 3 nodes", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(16)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphNodes(t, []int64{0, 10, 13}, sub.Value.Content.Nodes)
		})
		t.Run("The second node should be a graph with 2 edges", func(t *testing.T) {
			sub, err := sw.Collection.Graph.Node(16)
			if err != nil {
				t.Fatal(err)
			}

			dag.EnsureGraphEdges(t, map[int64][]int64{
				0:  {10},
				10: {13},
			}, sub.Value.Content.Edges)
		})
	})
}
