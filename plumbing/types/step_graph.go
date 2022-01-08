package types

import "errors"

var (
	ErrorAttachedNode = errors.New("node is attached to an existing node")
)

// A StepNode represents a single step.
// Sibling nodes are parallel steps.
// Child nodes are executed in the CI system after this one.
type StepNode struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *StepNode

	Step Step
}

func (n *StepNode) AppendChild(c *StepNode) error {
	if c.Parent != nil || c.PrevSibling != nil || c.NextSibling != nil {
		return ErrorAttachedNode
	}

	last := n.LastChild
	if last != nil {
		last.NextSibling = c
	} else {
		n.FirstChild = c
	}
	n.LastChild = c
	c.Parent = n
	c.PrevSibling = last

	return nil
}
