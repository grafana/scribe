package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/scribe/pipeline/dag"
	"github.com/grafana/scribe/state"
	"github.com/sirupsen/logrus"
)

var (
	ErrorNoSteps = errors.New("no steps were provided")
)

// WalkFunc is implemented by the pipeline 'Clients'. This function is executed for each Step.
type StepWalkFunc func(context.Context, Step) error

// PipelineWalkFunc is implemented by the pipeline 'Clients'. This function is executed for each pipeline.
// This function follows the same rules for pipelines as the StepWalker func does for pipelines. If multiple pipelines are provided in the steps argument,
// then those pipelines are intended to be executed in parallel.
type PipelineWalkFunc func(context.Context, ...Pipeline) error

// Walker is an interface that collections of steps should satisfy.
type Walker interface {
	WalkSteps(context.Context, int64, StepWalkFunc) error
	WalkPipelines(context.Context, PipelineWalkFunc) error
}

func StepIDs(steps []Step) []int64 {
	ids := make([]int64, len(steps))
	for i, v := range steps {
		ids[i] = v.ID
	}

	return ids
}

// Collection defines a directed acyclic Graph that stores a collection of Steps.
// When using Scribe with "scribe.New", this collection contains a graph with one pipeline in it.
// When using Scribe with "scribe.NewMulti", this collection contains a graph with several pipelines in it.
type Collection struct {
	Graph *dag.Graph[Pipeline]
}

func withoutBackgroundSteps(steps []Step) []Step {
	s := []Step{}

	for i, v := range steps {
		if v.Type != StepTypeBackground {
			s = append(s, steps[i])
		}
	}

	return s
}

// NodeListToSteps converts a list of Nodes to a list of Steps
func NodesToSteps(nodes []dag.Node[Step]) []Step {
	steps := make([]Step, len(nodes))

	for i, v := range nodes {
		steps[i] = v.Value
	}

	return steps
}

// AdjNodesToPipelines converts a list of Nodes (with type Pipeline) to a list of Pipelines
func AdjNodesToPipelines(nodes []*dag.Node[Pipeline]) []Pipeline {
	pipelines := make([]Pipeline, len(nodes))

	for i, v := range nodes {
		pipelines[i] = v.Value
	}

	return pipelines
}

// NodesToPipelines converts a list of Nodes (with type Pipeline) to a list of Pipelines
func NodesToPipelines(nodes []dag.Node[Pipeline]) []Pipeline {
	pipelines := make([]Pipeline, len(nodes))

	for i, v := range nodes {
		pipelines[i] = v.Value
	}

	return pipelines
}

func pipelinesEqual(a, b []Pipeline) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if b[i].ID != v.ID {
			return false
		}
	}

	return true
}

// BuildStepEdges generates the edges in each pipeline based on the required and provided args of each step.
func (c *Collection) BuildStepEdges(log logrus.FieldLogger, rootArgs ...state.Argument) error {
	for _, v := range c.Graph.Nodes {
		// default pipeline has no steps
		if v.ID == 0 {
			continue
		}

		if err := v.Value.BuildEdges(rootArgs...); err != nil {
			return err
		}

		log.WithFields(logrus.Fields{
			"pipeline": v.Value.Name,
			"nodes":    len(v.Value.Graph.Nodes),
			"edges":    len(v.Value.Graph.Edges),
		}).Debugln("Done building graph")
	}
	return nil
}

func (c *Collection) WalkSteps(ctx context.Context, pipelineID int64, wf StepWalkFunc) error {
	node, err := c.Graph.Node(pipelineID)
	if err != nil {
		return fmt.Errorf("could not find pipeline '%d'. %w", pipelineID, err)
	}

	pipeline := node.Value

	return pipeline.Graph.BreadthFirstSearch(0, func(n *dag.Node[Step]) error {
		if n.ID == 0 {
			return nil
		}

		return wf(ctx, n.Value)
	})
}

// AddEvents adds the list of events to the pipeline with 'pipelineID'.
// Events are not unique themselves are really only a list of arguments.
func (c *Collection) AddEvents(pipelineID int64, events ...Event) error {
	node, err := c.Graph.Node(pipelineID)
	if err != nil {
		return err
	}

	pipeline := node.Value
	pipeline.Events = events
	node.Value = pipeline
	return nil
}

// stepVisitFunc returns a dag.VisitFunc that popules the provided list of `steps` with the order that they should be ran.
func (c *Collection) pipelineVisitFunc(ctx context.Context, wf PipelineWalkFunc) dag.VisitFunc[Pipeline] {
	var (
		adj  = []Pipeline{}
		next = []Pipeline{}
	)

	return func(n *dag.Node[Pipeline]) error {
		if n.ID == 0 {
			adj = AdjNodesToPipelines(c.Graph.Adj(0))
			return nil
		}

		next = append(next, n.Value)

		if pipelinesEqual(adj, next) {
			if err := wf(ctx, next...); err != nil {
				return err
			}

			adj = AdjNodesToPipelines(c.Graph.Adj(n.ID))
			next = []Pipeline{}
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

// Add adds a new list of Steps to a pipeline. The order in which steps are added have no particular meaning.
func (c *Collection) AddSteps(pipelineID int64, steps ...Step) error {
	// Find the pipeline in our Graph of pipelines
	v, err := c.Graph.Node(pipelineID)
	if err != nil {
		return fmt.Errorf("error getting pipeline graph: %w", err)
	}

	if err := v.Value.AddSteps(steps...); err != nil {
		return fmt.Errorf("error adding steps to pipeline graph: %w", err)
	}

	// TODO: Should we do something here if steps.Type == StepTypeBackground?
	return nil
}

func (c *Collection) addPipeline(p Pipeline) error {
	if err := c.Graph.AddNode(p.ID, p); err != nil {
		return fmt.Errorf("error adding new pipeline to graph: %w", err)
	}

	if len(p.Dependencies) == 0 {
		if err := c.Graph.AddEdge(0, p.ID); err != nil {
			return err
		}
	}

	for _, v := range p.Dependencies {
		if err := c.Graph.AddEdge(v.ID, p.ID); err != nil {
			return err
		}
	}

	return nil
}

// AppendPipeline adds a populated sub-pipeline of Steps to the Graph.
func (c *Collection) AddPipelines(p ...Pipeline) error {
	for _, v := range p {
		if err := c.addPipeline(v); err != nil {
			return err
		}
	}
	return nil
}

// ByID should return the Step that corresponds with a specific ID
func (c *Collection) ByID(ctx context.Context, id int64) ([]Step, error) {
	steps := []Step{}

	// Search every pipeline and step for the listed IDs
	if err := c.WalkPipelines(ctx, func(ctx context.Context, pipelines ...Pipeline) error {
		for _, pipeline := range pipelines {
			for _, node := range pipeline.Graph.Nodes {
				if node.Value.ID == id {
					steps = []Step{node.Value}
					return dag.ErrorBreak
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(steps) == 0 {
		return nil, errors.New("no step found")
	}

	return steps, nil
}

// ByName should return the Step that corresponds with a specific Name
func (c *Collection) ByName(ctx context.Context, name string) ([]Step, error) {
	steps := []Step{}

	// Search every pipeline and step for the listed IDs
	if err := c.WalkPipelines(ctx, func(ctx context.Context, pipelines ...Pipeline) error {
		for _, pipeline := range pipelines {
			if pipeline.Name == name {
				// uhhh.. todo
				continue
			}
			for _, node := range pipeline.Graph.Nodes {
				if node.Value.Name == name {
					steps = []Step{node.Value}
					return dag.ErrorBreak
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return steps, nil
}

// PipelinesByName should return the Pipelines that corresponds with a specified names
func (c *Collection) PipelinesByName(ctx context.Context, names []string) ([]Pipeline, error) {
	var (
		retP  = make([]Pipeline, len(names))
		found = false
	)

	// Search every pipeline for the listed names
	if err := c.WalkPipelines(ctx, func(ctx context.Context, pipelines ...Pipeline) error {
		for i, argPipeline := range names {
			for _, pipeline := range pipelines {
				if strings.EqualFold(pipeline.Name, argPipeline) {
					pipeline.Dependencies = []Pipeline{}
					retP[i] = pipeline
					found = true
					break
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if !found {
		return nil, errors.New("no matching pipelines found")
	}
	return retP, nil
}

func (c *Collection) PipelinesByEvent(ctx context.Context, name string) ([]Pipeline, error) {
	ret := []Pipeline{}

	// Search every pipeline for the event
	if err := c.WalkPipelines(ctx, func(ctx context.Context, pipelines ...Pipeline) error {
		for _, pipeline := range pipelines {
			for _, event := range pipeline.Events {
				if event.Name == name {
					pipeline.Dependencies = []Pipeline{}
					ret = append(ret, pipeline)
					break
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if len(ret) == 0 {
		return nil, errors.New("no matching pipelines found")
	}

	return ret, nil
}

// NewCollectinoWithSteps creates a new Collection with a single pipeline from a list of Steps.
func NewCollectionWithSteps(pipelineName string, steps ...Step) (*Collection, error) {
	var (
		col       = NewCollection()
		id  int64 = 1
	)
	if err := col.AddPipelines(New(pipelineName, id)); err != nil {
		return nil, err
	}

	if err := col.AddSteps(id, steps...); err != nil {
		return nil, err
	}

	return col, nil
}

func NewCollection() *Collection {
	graph := dag.New[Pipeline]()
	graph.AddNode(0, New("default", 0))
	return &Collection{
		Graph: graph,
	}
}
