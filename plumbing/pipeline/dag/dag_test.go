package dag_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/grafana/scribe/plumbing/pipeline/dag"
	"github.com/grafana/scribe/plumbing/testutil"
)

func EnsureNodesExist[T any](t *testing.T, nodes []*dag.Node[T], ids ...int64) {
	if len(nodes) != len(ids) {
		t.Fatalf("Unequal slice lengths. Expected length '%d', but received length '%d'", len(ids), len(nodes))
	}

	exist := make([]int64, len(nodes))
	for i, v := range nodes {
		exist[i] = v.ID
	}
	if !testutil.Int64SlicesEqual(exist, ids) {
		t.Fatalf("Node list and expected not equal. Nodes: '%v'. Expected: '%v'", exist, ids)
	}
}

type Node struct{}

func TestGraphAddNode(t *testing.T) {
	t.Run("The graph should have 5 nodes if 4 nodes were added, as the root is added at index 0 by default", func(t *testing.T) {
		g := dag.New[Node]()

		testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(4, Node{}), nil)

		if len(g.Nodes) != 4 {
			t.Fatalf("Expected 4 nodes, found %d", len(g.Nodes))
		}
	})

	t.Run("AddNode should return an ErrorDuplicateID if a node is added with an ID that already exists", func(t *testing.T) {
		g := dag.New[Node]()

		testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(4, Node{}), dag.ErrorDuplicateID)
	})
}

func TestGraphAddEdge(t *testing.T) {
	t.Run("The graph should have 4 edges if 4 edges were added", func(t *testing.T) {
		g := dag.New[Node]()

		testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(5, Node{}), nil)

		testutil.EnsureError(t, g.AddEdge(1, 2), nil)
		testutil.EnsureError(t, g.AddEdge(2, 3), nil)
		testutil.EnsureError(t, g.AddEdge(3, 4), nil)
		testutil.EnsureError(t, g.AddEdge(4, 5), nil)

		if len(g.Edges) != 4 {
			t.Fatalf("Expected 4 edges, found %d", len(g.Edges))
		}
	})

	t.Run("AddEdge should return an ErrorNotFound an edge is added on a node that does not exist", func(t *testing.T) {
		g := dag.New[Node]()

		testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
		testutil.EnsureError(t, g.AddNode(3, Node{}), nil)

		testutil.EnsureError(t, g.AddEdge(1, 2), nil)
		testutil.EnsureError(t, g.AddEdge(2, 3), nil)
		testutil.EnsureError(t, g.AddEdge(3, 4), dag.ErrorNotFound)
		testutil.EnsureError(t, g.AddEdge(10, 10), dag.ErrorNotFound)
	})
}

func TestGraphAdj(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(5, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(6, Node{}), nil)

	testutil.EnsureError(t, g.AddEdge(1, 2), nil)
	testutil.EnsureError(t, g.AddEdge(1, 3), nil)
	testutil.EnsureError(t, g.AddEdge(1, 4), nil)
	testutil.EnsureError(t, g.AddEdge(3, 5), nil)
	testutil.EnsureError(t, g.AddEdge(3, 6), nil)

	t.Run("Adj should return nodes connected by an edge", func(t *testing.T) {
		adj := g.Adj(1)
		if len(adj) != 3 {
			t.Fatalf("Expected 3 adjacent nodes to node 1. Received '%d'", len(adj))
		}

		EnsureNodesExist(t, adj, 2, 3, 4)
	})

	t.Run("Adj should not return parent nodes that have edges to this node", func(t *testing.T) {
		adj := g.Adj(2)
		if adj != nil {
			t.Fatalf("Expected Adj to return nil, but instead received (len '%d') '%v'", len(adj), adj)
		}
	})
}

func TestGraphBreadthFirstSearch_12(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(0, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(5, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(6, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(7, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(8, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(9, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(10, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(11, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(12, Node{}), nil)

	testutil.EnsureError(t, g.AddEdge(0, 1), nil)
	testutil.EnsureError(t, g.AddEdge(1, 4), nil)
	testutil.EnsureError(t, g.AddEdge(4, 9), nil)
	testutil.EnsureError(t, g.AddEdge(4, 5), nil)
	testutil.EnsureError(t, g.AddEdge(5, 12), nil)
	testutil.EnsureError(t, g.AddEdge(12, 11), nil)
	testutil.EnsureError(t, g.AddEdge(11, 3), nil)
	testutil.EnsureError(t, g.AddEdge(11, 6), nil)
	testutil.EnsureError(t, g.AddEdge(11, 2), nil)
	testutil.EnsureError(t, g.AddEdge(2, 10), nil)
	testutil.EnsureError(t, g.AddEdge(10, 7), nil)
	testutil.EnsureError(t, g.AddEdge(7, 8), nil)

	var (
		expectedOrder = []int64{0, 1, 4, 9, 5, 12, 11, 3, 6, 2, 10, 7, 8}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	testutil.EnsureError(t, g.BreadthFirstSearch(0, visitFunc), nil)

	if !testutil.Int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}

	t.Run("if the visitfunc returns an error, BreadthFirstSearch should return that same error", func(t *testing.T) {
		err := errors.New("test error")
		vf := func(node *dag.Node[Node]) error {
			return err
		}

		testutil.EnsureError(t, g.BreadthFirstSearch(0, vf), err)
	})

	t.Run("if the visitfunc returns an error that is a 'dag.ErrorBreak',  BreadthFirstSearch should stop without returning an error", func(t *testing.T) {
		it := 0
		vf := func(node *dag.Node[Node]) error {
			it++
			return fmt.Errorf("example break error: %w", dag.ErrorBreak)
		}

		testutil.EnsureError(t, g.BreadthFirstSearch(0, vf), nil)
		if it != 1 {
			t.Fatalf("Expected 1 iteration before stopping, but found '%d'", it)
		}
	})

	t.Run("if an invalid starting point is provided, then BreadthFirstSearch should return an ErrorNotFound", func(t *testing.T) {
		testutil.EnsureError(t, g.BreadthFirstSearch(101, func(node *dag.Node[Node]) error { return nil }), dag.ErrorNotFound)
	})

	t.Run("if a nil visitFunc is provided, then BreadthFirstSearch should return an ErrorNoVisitFunc", func(t *testing.T) {
		testutil.EnsureError(t, g.BreadthFirstSearch(101, nil), dag.ErrorNoVisitFunc)
	})
}

func TestGraphDepthFirstSearch_12(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(0, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(5, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(6, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(7, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(8, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(9, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(10, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(11, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(12, Node{}), nil)

	testutil.EnsureError(t, g.AddEdge(0, 1), nil)
	testutil.EnsureError(t, g.AddEdge(1, 4), nil)
	testutil.EnsureError(t, g.AddEdge(4, 9), nil)
	testutil.EnsureError(t, g.AddEdge(4, 5), nil)
	testutil.EnsureError(t, g.AddEdge(5, 12), nil)
	testutil.EnsureError(t, g.AddEdge(12, 11), nil)
	testutil.EnsureError(t, g.AddEdge(11, 3), nil)
	testutil.EnsureError(t, g.AddEdge(11, 6), nil)
	testutil.EnsureError(t, g.AddEdge(11, 2), nil)
	testutil.EnsureError(t, g.AddEdge(2, 10), nil)
	testutil.EnsureError(t, g.AddEdge(10, 7), nil)
	testutil.EnsureError(t, g.AddEdge(7, 8), nil)

	var (
		expectedOrder = []int64{0, 1, 4, 9, 5, 12, 11, 3, 6, 2, 10, 7, 8}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	testutil.EnsureError(t, g.DepthFirstSearch(0, visitFunc), nil)

	if !testutil.Int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}

	t.Run("if the visitfunc returns an error, DepthFirstSearch should return that same error", func(t *testing.T) {
		err := errors.New("test error")
		vf := func(node *dag.Node[Node]) error {
			return err
		}

		testutil.EnsureError(t, g.DepthFirstSearch(0, vf), err)
	})

	t.Run("if the visitfunc returns an error that is a 'dag.ErrorBreak',  DepthFirstSearch should stop without returning an error", func(t *testing.T) {
		it := 0
		vf := func(node *dag.Node[Node]) error {
			it++
			return fmt.Errorf("example break error: %w", dag.ErrorBreak)
		}

		testutil.EnsureError(t, g.DepthFirstSearch(0, vf), nil)
		if it != 1 {
			t.Fatalf("Expected 1 iteration before stopping, but found '%d'", it)
		}
	})

	t.Run("if an invalid starting point is provided, then DepthFirstSearch should return an ErrorNotFound", func(t *testing.T) {
		testutil.EnsureError(t, g.DepthFirstSearch(101, func(node *dag.Node[Node]) error { return nil }), dag.ErrorNotFound)
	})

	t.Run("if a nil visitFunc is provided, then DepthFirstSearch should return an ErrorNoVisitFunc", func(t *testing.T) {
		testutil.EnsureError(t, g.DepthFirstSearch(101, nil), dag.ErrorNoVisitFunc)
	})
}

func TestGraphDepthFirstSearch_24(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(0, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(5, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(6, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(7, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(8, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(9, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(10, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(11, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(12, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(13, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(14, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(15, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(16, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(17, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(18, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(19, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(20, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(21, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(22, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(23, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(24, Node{}), nil)

	testutil.EnsureError(t, g.AddEdge(0, 1), nil)
	testutil.EnsureError(t, g.AddEdge(1, 4), nil)
	testutil.EnsureError(t, g.AddEdge(4, 9), nil)
	testutil.EnsureError(t, g.AddEdge(4, 5), nil)
	testutil.EnsureError(t, g.AddEdge(5, 12), nil)
	testutil.EnsureError(t, g.AddEdge(12, 11), nil)
	testutil.EnsureError(t, g.AddEdge(11, 3), nil)
	testutil.EnsureError(t, g.AddEdge(11, 6), nil)
	testutil.EnsureError(t, g.AddEdge(11, 2), nil)
	testutil.EnsureError(t, g.AddEdge(2, 10), nil)
	testutil.EnsureError(t, g.AddEdge(10, 7), nil)
	testutil.EnsureError(t, g.AddEdge(7, 8), nil)
	testutil.EnsureError(t, g.AddEdge(8, 20), nil)
	testutil.EnsureError(t, g.AddEdge(8, 16), nil)
	testutil.EnsureError(t, g.AddEdge(8, 23), nil)
	testutil.EnsureError(t, g.AddEdge(8, 21), nil)
	testutil.EnsureError(t, g.AddEdge(8, 18), nil)
	testutil.EnsureError(t, g.AddEdge(18, 13), nil)
	testutil.EnsureError(t, g.AddEdge(13, 22), nil)
	testutil.EnsureError(t, g.AddEdge(22, 17), nil)
	testutil.EnsureError(t, g.AddEdge(17, 14), nil)
	testutil.EnsureError(t, g.AddEdge(17, 24), nil)
	testutil.EnsureError(t, g.AddEdge(17, 15), nil)
	testutil.EnsureError(t, g.AddEdge(17, 19), nil)

	var (
		expectedOrder = []int64{0, 1, 4, 9, 5, 12, 11, 3, 6, 2, 10, 7, 8, 20, 16, 23, 21, 18, 13, 22, 17, 14, 24, 15, 19}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	testutil.EnsureError(t, g.DepthFirstSearch(0, visitFunc), nil)

	if !testutil.Int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}
}

func TestDepthFirstSearchUnoptimized(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(0, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(5, Node{}), nil)

	testutil.EnsureError(t, g.AddEdge(0, 1), nil)
	testutil.EnsureError(t, g.AddEdge(1, 2), nil)
	testutil.EnsureError(t, g.AddEdge(1, 3), nil)
	testutil.EnsureError(t, g.AddEdge(2, 5), nil)
	testutil.EnsureError(t, g.AddEdge(3, 5), nil)
	testutil.EnsureError(t, g.AddEdge(2, 4), nil)
	testutil.EnsureError(t, g.AddEdge(3, 4), nil)

	var (
		expectedOrder = []int64{0, 1, 2, 5, 4, 3}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	testutil.EnsureError(t, g.DepthFirstSearch(0, visitFunc), nil)

	if !testutil.Int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}
}

func TestGraphNode(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)

	t.Run("graph.Node should return the node with a specific ID", func(t *testing.T) {
		node, err := g.Node(4)
		if err != nil {
			t.Fatalf("Expected no error but got '%s'", err.Error())
		}

		if node.ID != 4 {
			t.Fatalf("Expected node with id 4, got '%d'", node.ID)
		}
	})

	t.Run("graph.Node should return an ErrorNotFound if no node was found", func(t *testing.T) {
		_, err := g.Node(5)
		if err == nil {
			t.Fatal("Expected an error but did not receive one")
		}

		if !errors.Is(err, dag.ErrorNotFound) {
			t.Fatal("Expected error to be a dag.ErrorNotFound but got:", err.Error())
		}
	})
}

func TestGraphNodeLIst(t *testing.T) {
	g := dag.New[Node]()

	testutil.EnsureError(t, g.AddNode(1, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(2, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(3, Node{}), nil)
	testutil.EnsureError(t, g.AddNode(4, Node{}), nil)

	t.Run("graph.Node should return the nodes with the specified IDs", func(t *testing.T) {
		nodes, err := g.NodeList(1, 2, 3, 4)
		if err != nil {
			t.Fatalf("Expected no error but got '%s'", err.Error())
		}

		EnsureNodesExist(t, nodes, 1, 2, 3, 4)
	})

	t.Run("graph.Node should return an ErrorNotFound if no node was found", func(t *testing.T) {
		_, err := g.NodeList(2, 3, 5)
		if err == nil {
			t.Fatal("Expected an error but did not receive one")
		}

		if !errors.Is(err, dag.ErrorNotFound) {
			t.Fatal("Expected error to be a dag.ErrorNotFound but got:", err.Error())
		}
	})
}
