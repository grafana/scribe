package dag_test

import (
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/grafana/shipwright/plumbing/pipeline/dag"
)

func EnsureError(t *testing.T, err, expect error) {
	if expect == nil && err != nil {
		t.Fatalf("Expected no error but received '%s'", err.Error())
	}

	if !errors.Is(err, expect) {
		t.Fatalf("Expected error '%s' but received '%s'", expect.Error(), err.Error())
	}
}

func int64SlicesEqual(a []int64, b []int64) bool {
	sort.Slice(a, int64SortFunc(a))
	sort.Slice(b, int64SortFunc(b))

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func int64SortFunc(int64s []int64) func(i, j int) bool {
	return func(i, j int) bool {
		return int64s[i] < int64s[j]
	}
}

func EnsureNodesExist[T any](t *testing.T, nodes []*dag.Node[T], ids ...int64) {
	if len(nodes) != len(ids) {
		t.Fatalf("Unequal slice lengths. Expected length '%d', but received length '%d'", len(ids), len(nodes))
	}

	exist := make([]int64, len(nodes))
	for i, v := range nodes {
		exist[i] = v.ID
	}
	if !int64SlicesEqual(exist, ids) {
		t.Fatalf("Node list and expected not equal. Nodes: '%v'. Expected: '%v'", exist, ids)
	}
}

type Node struct{}

func TestGraphAddNode(t *testing.T) {
	t.Run("The graph should have 5 nodes if 4 nodes were added, as the root is added at index 0 by default", func(t *testing.T) {
		g := dag.New[Node]()

		EnsureError(t, g.AddNode(1, Node{}), nil)
		EnsureError(t, g.AddNode(2, Node{}), nil)
		EnsureError(t, g.AddNode(3, Node{}), nil)
		EnsureError(t, g.AddNode(4, Node{}), nil)

		if len(g.Nodes) != 5 {
			t.Fatalf("Expected 5 nodes, found %d", len(g.Nodes))
		}
	})

	t.Run("AddNode should return an ErrorDuplicateID if a node is added with an ID that already exists", func(t *testing.T) {
		g := dag.New[Node]()

		EnsureError(t, g.AddNode(1, Node{}), nil)
		EnsureError(t, g.AddNode(2, Node{}), nil)
		EnsureError(t, g.AddNode(3, Node{}), nil)
		EnsureError(t, g.AddNode(4, Node{}), nil)
		EnsureError(t, g.AddNode(4, Node{}), dag.ErrorDuplicateID)
	})

	t.Run("A node can not be added at index 0", func(t *testing.T) {
		g := dag.New[Node]()
		EnsureError(t, g.AddNode(0, Node{}), dag.ErrorDuplicateID)
	})
}

func TestGraphAddEdge(t *testing.T) {
	t.Run("The graph should have 4 edges if 4 edges were added", func(t *testing.T) {
		g := dag.New[Node]()

		EnsureError(t, g.AddNode(1, Node{}), nil)
		EnsureError(t, g.AddNode(2, Node{}), nil)
		EnsureError(t, g.AddNode(3, Node{}), nil)
		EnsureError(t, g.AddNode(4, Node{}), nil)
		EnsureError(t, g.AddNode(5, Node{}), nil)

		EnsureError(t, g.AddEdge(1, 2), nil)
		EnsureError(t, g.AddEdge(2, 3), nil)
		EnsureError(t, g.AddEdge(3, 4), nil)
		EnsureError(t, g.AddEdge(4, 5), nil)

		if len(g.Edges) != 4 {
			t.Fatalf("Expected 4 edges, found %d", len(g.Edges))
		}
	})

	t.Run("AddEdge should return an ErrorNotFound an edge is added on a node that does not exist", func(t *testing.T) {
		g := dag.New[Node]()

		EnsureError(t, g.AddNode(1, Node{}), nil)
		EnsureError(t, g.AddNode(2, Node{}), nil)
		EnsureError(t, g.AddNode(3, Node{}), nil)

		EnsureError(t, g.AddEdge(1, 2), nil)
		EnsureError(t, g.AddEdge(2, 3), nil)
		EnsureError(t, g.AddEdge(3, 4), dag.ErrorNotFound)
		EnsureError(t, g.AddEdge(10, 10), dag.ErrorNotFound)
	})
}

func TestGraphAdj(t *testing.T) {
	g := dag.New[Node]()

	EnsureError(t, g.AddNode(1, Node{}), nil)
	EnsureError(t, g.AddNode(2, Node{}), nil)
	EnsureError(t, g.AddNode(3, Node{}), nil)
	EnsureError(t, g.AddNode(4, Node{}), nil)
	EnsureError(t, g.AddNode(5, Node{}), nil)
	EnsureError(t, g.AddNode(6, Node{}), nil)

	EnsureError(t, g.AddEdge(1, 2), nil)
	EnsureError(t, g.AddEdge(1, 3), nil)
	EnsureError(t, g.AddEdge(1, 4), nil)
	EnsureError(t, g.AddEdge(3, 5), nil)
	EnsureError(t, g.AddEdge(3, 6), nil)

	t.Run("Adj should return nodes connected by an edge", func(t *testing.T) {
		adj := g.Adj(1)
		if len(adj) != 3 {
			t.Fatalf("Expected 3 adjacent nodes to node 1. Received '%d'", len(adj))
		}

		EnsureNodesExist(t, adj, 2, 3, 4)
	})

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

func TestGraphDepthFirstSearch_12(t *testing.T) {
	g := dag.New[Node]()

	EnsureError(t, g.AddNode(1, Node{}), nil)
	EnsureError(t, g.AddNode(2, Node{}), nil)
	EnsureError(t, g.AddNode(3, Node{}), nil)
	EnsureError(t, g.AddNode(4, Node{}), nil)
	EnsureError(t, g.AddNode(5, Node{}), nil)
	EnsureError(t, g.AddNode(6, Node{}), nil)
	EnsureError(t, g.AddNode(7, Node{}), nil)
	EnsureError(t, g.AddNode(8, Node{}), nil)
	EnsureError(t, g.AddNode(9, Node{}), nil)
	EnsureError(t, g.AddNode(10, Node{}), nil)
	EnsureError(t, g.AddNode(11, Node{}), nil)
	EnsureError(t, g.AddNode(12, Node{}), nil)

	EnsureError(t, g.AddEdge(0, 1), nil)
	EnsureError(t, g.AddEdge(1, 4), nil)
	EnsureError(t, g.AddEdge(4, 9), nil)
	EnsureError(t, g.AddEdge(4, 5), nil)
	EnsureError(t, g.AddEdge(5, 12), nil)
	EnsureError(t, g.AddEdge(12, 11), nil)
	EnsureError(t, g.AddEdge(11, 3), nil)
	EnsureError(t, g.AddEdge(11, 6), nil)
	EnsureError(t, g.AddEdge(11, 2), nil)
	EnsureError(t, g.AddEdge(2, 10), nil)
	EnsureError(t, g.AddEdge(10, 7), nil)
	EnsureError(t, g.AddEdge(7, 8), nil)

	var (
		expectedOrder = []int64{0, 1, 4, 5, 9, 12, 11, 2, 3, 6, 10, 7, 8}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	EnsureError(t, g.DepthFirstSearch(0, visitFunc), nil)

	if !int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}

	t.Run("if the visitfunc returns an error, DepthFirstSearch should return that same error", func(t *testing.T) {
		err := errors.New("test error")
		vf := func(node *dag.Node[Node]) error {
			return err
		}

		EnsureError(t, g.DepthFirstSearch(0, vf), err)
	})

	t.Run("if the visitfunc returns an error that is a 'dag.ErrorBreak',  DepthFirstSearch should stop without returning an error", func(t *testing.T) {
		it := 0
		vf := func(node *dag.Node[Node]) error {
			it++
			return fmt.Errorf("example break error: %w", dag.ErrorBreak)
		}

		EnsureError(t, g.DepthFirstSearch(0, vf), nil)
		if it != 1 {
			t.Fatalf("Expected 1 iteration before stopping, but found '%d'", it)
		}
	})
}

func TestGraphDepthFirstSearch_24(t *testing.T) {
	g := dag.New[Node]()

	EnsureError(t, g.AddNode(1, Node{}), nil)
	EnsureError(t, g.AddNode(2, Node{}), nil)
	EnsureError(t, g.AddNode(3, Node{}), nil)
	EnsureError(t, g.AddNode(4, Node{}), nil)
	EnsureError(t, g.AddNode(5, Node{}), nil)
	EnsureError(t, g.AddNode(6, Node{}), nil)
	EnsureError(t, g.AddNode(7, Node{}), nil)
	EnsureError(t, g.AddNode(8, Node{}), nil)
	EnsureError(t, g.AddNode(9, Node{}), nil)
	EnsureError(t, g.AddNode(10, Node{}), nil)
	EnsureError(t, g.AddNode(11, Node{}), nil)
	EnsureError(t, g.AddNode(12, Node{}), nil)
	EnsureError(t, g.AddNode(13, Node{}), nil)
	EnsureError(t, g.AddNode(14, Node{}), nil)
	EnsureError(t, g.AddNode(15, Node{}), nil)
	EnsureError(t, g.AddNode(16, Node{}), nil)
	EnsureError(t, g.AddNode(17, Node{}), nil)
	EnsureError(t, g.AddNode(18, Node{}), nil)
	EnsureError(t, g.AddNode(19, Node{}), nil)
	EnsureError(t, g.AddNode(20, Node{}), nil)
	EnsureError(t, g.AddNode(21, Node{}), nil)
	EnsureError(t, g.AddNode(22, Node{}), nil)
	EnsureError(t, g.AddNode(23, Node{}), nil)
	EnsureError(t, g.AddNode(24, Node{}), nil)

	EnsureError(t, g.AddEdge(0, 1), nil)
	EnsureError(t, g.AddEdge(1, 4), nil)
	EnsureError(t, g.AddEdge(4, 9), nil)
	EnsureError(t, g.AddEdge(4, 5), nil)
	EnsureError(t, g.AddEdge(5, 12), nil)
	EnsureError(t, g.AddEdge(12, 11), nil)
	EnsureError(t, g.AddEdge(11, 3), nil)
	EnsureError(t, g.AddEdge(11, 6), nil)
	EnsureError(t, g.AddEdge(11, 2), nil)
	EnsureError(t, g.AddEdge(2, 10), nil)
	EnsureError(t, g.AddEdge(10, 7), nil)
	EnsureError(t, g.AddEdge(7, 8), nil)
	EnsureError(t, g.AddEdge(8, 20), nil)
	EnsureError(t, g.AddEdge(8, 16), nil)
	EnsureError(t, g.AddEdge(8, 23), nil)
	EnsureError(t, g.AddEdge(8, 21), nil)
	EnsureError(t, g.AddEdge(8, 18), nil)
	EnsureError(t, g.AddEdge(18, 13), nil)
	EnsureError(t, g.AddEdge(13, 22), nil)
	EnsureError(t, g.AddEdge(22, 17), nil)
	EnsureError(t, g.AddEdge(17, 14), nil)
	EnsureError(t, g.AddEdge(17, 24), nil)
	EnsureError(t, g.AddEdge(17, 15), nil)
	EnsureError(t, g.AddEdge(17, 19), nil)

	var (
		expectedOrder = []int64{0, 1, 4, 5, 9, 12, 11, 2, 3, 6, 10, 7, 8, 16, 20, 21, 23, 18, 13, 22, 17, 14, 15, 19, 24}
		order         = []int64{}
	)

	var visitFunc = func(node *dag.Node[Node]) error {
		order = append(order, node.ID)
		return nil
	}

	EnsureError(t, g.DepthFirstSearch(0, visitFunc), nil)

	if !int64SlicesEqual(expectedOrder, order) {
		t.Fatalf("Nodes visited in unexpected order. Expected: '%v', visited order: '%v'", expectedOrder, order)
	}
}

func TestGraphNode(t *testing.T) {
	g := dag.New[Node]()

	EnsureError(t, g.AddNode(1, Node{}), nil)
	EnsureError(t, g.AddNode(2, Node{}), nil)
	EnsureError(t, g.AddNode(3, Node{}), nil)
	EnsureError(t, g.AddNode(4, Node{}), nil)

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
