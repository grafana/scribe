package types

import (
	"errors"
)

var (
	ErrorAttachedNode = errors.New("node is attached to an existing node")
)

// A Node represents a single step.
// Sibling nodes are parallel steps.
// Child nodes are executed in the CI system after this one.
type Node struct {
	FirstChild, LastChild    *Node
	PrevSibling, NextSibling *Node

	Step Step
}

func NewNode(step Step) *Node {
	return &Node{
		Step: step,
	}
}

// Collection stores Steps for creating a pipeline
type Collection interface {
	// Append appends the list of steps as if each step.
	// The individual steps themselves should have information on their dependent steps, which should determine
	// where in the tree they should live.
	Append(steps ...Step) error

	// AppendLineage appends the list of steps as if each one is a child of the previous one.
	AppendLineage(steps ...Step) error

	Walk(func(Step) error)
}
