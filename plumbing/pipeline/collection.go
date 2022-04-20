package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/grafana/shipwright/plumbing/pipeline/dag"
)

var (
	ErrorNoSteps = errors.New("no steps were provided")
)

func NewPipelineNode(name string, id int64) Step[Pipeline] {
	return Step[Pipeline]{
		Name:    name,
		Serial:  id,
		Content: NewPipeline(),
	}
}

func NewPipeline() Pipeline {
	graph := dag.New[Step[StepList]]()
	graph.AddNode(0, Step[StepList]{})

	return Pipeline{
		Graph: graph,
	}
}

// AddStep adds the steps as a single node in the pipeline.
func (p *Pipeline) AddSteps(steps Step[StepList]) error {
	if err := p.AddNode(steps.Serial, steps); err != nil {
		return err
	}

	return nil
}

func nodeID(steps []Step[Action]) int64 {
	return steps[len(steps)-1].Serial
}

// WalkFunc is implemented by the pipeline 'Clients'. This function is executed for each step.
// If multiple steps are provided in the argument, then they were provided in "Parallel".
// If one step in the list of steps is of type "Background", then they all should be.
type StepWalkFunc func(context.Context, ...Step[Action]) error

// PipelineWalkFunc is implemented by the pipeline 'Clients'. This function is executed for each pipeline.
// This function follows the same rules for pipelines as the StepWalker func does for pipelines. If multiple pipelines are provided in the steps argument,
// then those pipelines are intended to be executed in parallel.
type PipelineWalkFunc func(context.Context, ...Step[Pipeline]) error

// Walker is an interface that collections of steps should satisfy.
type Walker interface {
	WalkSteps(context.Context, int64, StepWalkFunc) error
	WalkPipelines(context.Context, PipelineWalkFunc) error
}

func StepIDs(steps []Step[Action]) []int64 {
	ids := make([]int64, len(steps))
	for i, v := range steps {
		ids[i] = v.Serial
	}

	return ids
}

// Collection defines a directed acyclic Graph that stores a collection of Steps.
type Collection struct {
	Graph *dag.Graph[Step[Pipeline]]
}

func stepListEqual(a, b []Step[Action]) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Serial != b[i].Serial {
			return false
		}
	}

	return true
}

func StepNames[T StepContent](s []Step[T]) []string {
	v := make([]string, len(s))
	for i := range s {
		v[i] = s[i].Name
	}

	return v
}

func withoutBackgroundSteps(steps []Step[Action]) []Step[Action] {
	s := []Step[Action]{}

	for i, v := range steps {
		if v.Type != StepTypeBackground {
			s = append(s, steps[i])
		}
	}

	return s
}

func AdjListToSteps[T StepContent](nodes []*dag.Node[Step[T]]) []Step[T] {
	steps := make([]Step[T], len(nodes))

	for i, v := range nodes {
		steps[i] = v.Value
	}

	return steps
}

func NodeListToSteps[T StepContent](nodes []dag.Node[Step[T]]) []Step[T] {
	steps := make([]Step[T], len(nodes))

	for i, v := range nodes {
		steps[i] = v.Value
	}

	return steps
}

func nodeListsEqual[T StepContent](a, b []Step[T]) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if b[i].Serial != v.Serial {
			return false
		}
	}

	return true
}

// stepVisitFunc returns a dag.VisitFunc that popules the provided list of `steps` with the order that they should be ran.
func (c *Collection) stepVisitFunc(ctx context.Context, wf StepWalkFunc) dag.VisitFunc[Step[StepList]] {
	return func(n *dag.Node[Step[StepList]]) error {
		if n.ID == 0 {
			return nil
		}

		list := n.Value

		// Because every group of steps run in parallel, they all share dependencies.
		// Those dependencies however should not be the single ID that represents the group,
		// but all of the steps that are contained within the group.
		deps := []Step[Action]{}
		for _, step := range list.Dependencies {
			for _, v := range step.Content {
				deps = append(deps, v)
			}
		}

		for i := range list.Content {
			list.Content[i].Dependencies = deps
		}
		return wf(ctx, list.Content...)
	}
}

func (c *Collection) WalkSteps(ctx context.Context, pipelineID int64, wf StepWalkFunc) error {
	node, err := c.Graph.Node(pipelineID)
	if err != nil {
		return fmt.Errorf("could not find pipeline '%d'. %w", pipelineID, err)
	}

	pipeline := node.Value

	if err := pipeline.Content.BreadthFirstSearch(0, c.stepVisitFunc(ctx, wf)); err != nil {
		return err
	}

	return nil
}

func (c *Collection) AddEvents(pipelineID int64, events ...Event) error {
	node, err := c.Graph.Node(pipelineID)
	if err != nil {
		return err
	}
	node.Value.Content.Events = append(node.Value.Content.Events, events...)

	return nil
}

// stepVisitFunc returns a dag.VisitFunc that popules the provided list of `steps` with the order that they should be ran.
func (c *Collection) pipelineVisitFunc(ctx context.Context, wf PipelineWalkFunc) dag.VisitFunc[Step[Pipeline]] {
	var (
		adj  = []Step[Pipeline]{}
		next = []Step[Pipeline]{}
	)

	return func(n *dag.Node[Step[Pipeline]]) error {
		log.Println("Visiting pipeline", n.ID, n.Value.Name, "nodes:", len(n.Value.Content.Nodes))
		if n.ID == 0 {
			adj = AdjListToSteps(c.Graph.Adj(0))
			return nil
		}

		next = append(next, n.Value)

		if nodeListsEqual(adj, next) {
			if err := wf(ctx, next...); err != nil {
				return err
			}

			adj = AdjListToSteps(c.Graph.Adj(n.ID))
			next = []Step[Pipeline]{}
		}

		return nil
	}
}

func (c *Collection) WalkPipelines(ctx context.Context, wf PipelineWalkFunc) error {
	if err := c.Graph.BreadthFirstSearch(0, c.pipelineVisitFunc(ctx, wf)); err != nil {
		return err
	}
	return nil
}

// Add adds a new list of Steps which are siblings to a pipeline.
// Because they are siblings, they must all depend on the same step(s).
func (c *Collection) AddSteps(pipelineID int64, steps Step[StepList]) error {
	// Find the pipeline in our Graph of pipelines
	v, err := c.Graph.Node(pipelineID)
	if err != nil {
		return fmt.Errorf("error getting pipeline graph: %w", err)
	}

	pipeline := v.Value.Content
	if err := pipeline.AddSteps(steps); err != nil {
		return fmt.Errorf("error adding steps to pipelien graph: %w", err)
	}

	// Background steps should only have an edge from the root node. This is automatically added as Background Steps do not have dependencies.
	// Because Backgorund steps are intended to persist until the pipeline terminates, they can't have child steps.
	if len(steps.Dependencies) == 0 {
		pipeline.AddEdge(0, steps.Serial)
	}

	if steps.Type == StepTypeBackground {
		return nil
	}

	for _, parent := range steps.Dependencies {
		if err := pipeline.AddEdge(parent.Serial, steps.Serial); err != nil {
			return fmt.Errorf("error adding edges to pipeline graph: %w", err)
		}
	}

	return nil
}

func (c *Collection) addPipeline(p Step[Pipeline]) error {
	if err := c.Graph.AddNode(p.Serial, p); err != nil {
		return fmt.Errorf("error adding new pipeline to graph: %w", err)
	}

	if len(p.Dependencies) == 0 {
		if err := c.Graph.AddEdge(0, p.Serial); err != nil {
			return err
		}
	}

	for _, v := range p.Dependencies {
		if err := c.Graph.AddEdge(v.Serial, p.Serial); err != nil {
			return err
		}
	}

	return nil
}

// AppendPipeline adds a populated sub-pipeline of Steps to the Graph.
func (c *Collection) AddPipelines(p ...Step[Pipeline]) error {
	for _, v := range p {
		if err := c.addPipeline(v); err != nil {
			return err
		}
	}
	return nil
}

// BySerial should return the Step that corresponds with a specific Serial
func (c *Collection) BySerial(int) (Step[Action], error) {
	return Step[Action]{}, nil
}

// ByName should return the Step that corresponds with a specific Name
func (c *Collection) ByName(string) (Step[Action], error) {
	return Step[Action]{}, nil
}

// Pipeline should return the pipeline that corresponds to a specific name
func (c *Collection) Pipeline(string) (Step[StepList], error) {
	return Step[StepList]{}, nil
}

// Sub creates a new Collection of the same type from a list of Steps
func (c *Collection) Sub(...Step[Action]) *Collection {
	return nil
}

func NewCollection() *Collection {
	graph := dag.New[Step[Pipeline]]()
	graph.AddNode(0, NewPipelineNode("default", 0))
	return &Collection{
		Graph: graph,
	}
}
