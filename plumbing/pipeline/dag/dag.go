package dag

import (
	"errors"
	"fmt"
)

var (
	ErrorDuplicateID = errors.New("node with ID already exists")
	ErrorNotFound    = errors.New("node with ID not found")
)

type Node[T any] struct {
	ID    int64
	Value T
}

type Edge[T any] struct {
	From *Node[T]
	To   *Node[T]
}

type Graph[T any] struct {
	Nodes []Node[T]
	Edges map[int64][]Edge[T]

	visited map[int64]bool
}

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

func (g *Graph[T]) dfs(id int64, visitFunc func(n *Node[T])) {
	g.visited[id] = true
	node, err := g.Node(id)
	if err != nil {
		panic(err)
	}

	visitFunc(&node)

	adj := g.Adj(id)
	if adj == nil {
		return
	}

	for _, v := range adj {
		g.dfs(v.ID, visitFunc)
	}
}

func (g *Graph[T]) DepthFirstSearch(start int64, visitFunc func(n *Node[T])) {
	g.visited = make(map[int64]bool, len(g.Nodes))
	g.dfs(start, visitFunc)
}

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
