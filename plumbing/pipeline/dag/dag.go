package dag

import (
	"errors"
	"fmt"
)

var (
	ErrorDuplicateID = errors.New("node with ID already exists")
	ErrorNotFound    = errors.New("node with ID not found")

	ErrorBreak = errors.New("break will stop the depth first search without an error")
)

// Node is a graph node that has an ID and data.
// Nodes are connected by Edges.
type Node[T any] struct {
	ID    int64
	Value T
}

// Edge is a connection from one node to another.
// Because this is a Directed graph, the edge has a direction.
// A connection from 'node A' to 'node B' is not the same as a connection from 'node B' to 'node A'.
type Edge[T any] struct {
	From *Node[T]
	To   *Node[T]
}

// Graph is a data structure that stores a list of Nodes (data) and Edges that connect nodes.
// Because it is a Directed graph, the edges connect from a node to another node, and the connection is not equal if reversed.
// Because it is an Acyclic graph, the nodes can not be connected in a loop or a cycle. If the nodes/edges look like (0 -> 1 -> 2 -> 0), then that is a cycle and is not allowed.
type Graph[T any] struct {
	Nodes []Node[T]
	Edges map[int64][]Edge[T]

	visited map[int64]bool
}

// AddNode adds a new node to the graph with the given ID and data (v).
func (g *Graph[T]) AddNode(id int64, v T) error {
	node := Node[T]{
		ID:    id,
		Value: v,
	}

	for _, v := range g.Nodes {
		if v.ID == node.ID {
			return fmt.Errorf("%w. id: %d", ErrorDuplicateID, node.ID)
		}
	}

	g.Nodes = append(g.Nodes, node)
	return nil
}

// AddEdge adds a new node from node with the ID 'from' to the node with the ID 'to'.
func (g *Graph[T]) AddEdge(from, to int64) error {
	var fromNode, toNode *Node[T]

	for i, v := range g.Nodes {
		if v.ID == from {
			fromNode = &g.Nodes[i]
		}
		if v.ID == to {
			toNode = &g.Nodes[i]
		}
		if fromNode != nil && toNode != nil {
			break
		}
	}

	if fromNode == nil {
		return fmt.Errorf("%w. id: %d", ErrorNotFound, from)
	}

	if toNode == nil {
		return fmt.Errorf("%w. id: %d", ErrorNotFound, to)
	}
	edges := g.Edges[from]
	g.Edges[from] = append(edges, Edge[T]{
		From: fromNode,
		To:   toNode,
	})

	return nil
}

// Node returns the node with the given ID.
// If no node is found, ErrorNotFound is returned.
func (g *Graph[T]) Node(id int64) (Node[T], error) {
	for _, v := range g.Nodes {
		if v.ID == id {
			return v, nil
		}
	}

	return Node[T]{}, ErrorNotFound
}

// Adj returns nodes with edges that start at the provided node (n) (Where 'From' is this node).
// This function does not return nodes with edges that end at the provided node (where 'To' is this node).
func (g *Graph[T]) Adj(id int64) []*Node[T] {
	edges, ok := g.Edges[id]
	if !ok {
		return nil
	}

	siblings := make([]*Node[T], len(edges))
	for i := range edges {
		siblings[i] = edges[i].To
	}

	return siblings
}

type VisitFunc[T any] func(n *Node[T]) error

func (g *Graph[T]) dfs(id int64, visitFunc VisitFunc[T]) error {
	g.visited[id] = true
	node, err := g.Node(id)
	if err != nil {
		panic(err)
	}

	if err := visitFunc(&node); err != nil {
		if errors.Is(err, ErrorBreak) {
			return nil
		}

		return err
	}

	adj := g.Adj(id)
	if adj == nil {
		return nil
	}

	for _, v := range adj {
		g.dfs(v.ID, visitFunc)
	}

	return nil
}

// DepthFirstSearch performs a depth-first search and calls the provided visitFunc callback for every node.
// 'visitFunc' is not called more than once per node.
// If 'visitFunc' returns an error, then so will this function.
// If 'visitFunc' returns ErrorBreak, then this function will return nil and will stop visiting nodes.
func (g *Graph[T]) DepthFirstSearch(start int64, visitFunc VisitFunc[T]) error {
	g.visited = make(map[int64]bool, len(g.Nodes))
	return g.dfs(start, visitFunc)
}

// New creates a new Graph with nodes that contain data with type T.
func New[T any]() *Graph[T] {
	return &Graph[T]{
		Nodes: []Node[T]{
			{
				ID: 0,
			},
		},
		Edges: map[int64][]Edge[T]{},
	}
}
