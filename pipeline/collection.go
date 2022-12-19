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
	ErrorNoPipelineProvider = errors.New("no pipeline in the graph provides a required argument")
	ErrorNoSteps            = errors.New("no steps were provided")
)

// Collection defines a directed acyclic Graph that stores a collection of Steps.
// When using Scribe with "scribe.New", this collection contains a graph with one pipeline in it.
// When using Scribe with "scribe.NewMulti", this collection contains a graph with several pipelines in it.
type Collection struct {
	Graph     *dag.Graph[Pipeline]
	Providers map[state.Argument]int64
	Root      []int64
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

	//col.Root = []int64{id}
	//log.Println("NewCollectionWithSteps: buiding graph edges")
	//log.Println("NewCollectionWithSteps: buiding graph edges")
	//if err := col.BuildEdges(logrus.StandardLogger(), ClientProvidedArguments...); err != nil {
	//	return nil, err
	//}
	//log.Println("NewCollectionWithSteps: done buiding graph edges")
	//log.Println("NewCollectionWithSteps: done buiding graph edges")

	return col, nil
}

func NewCollection() *Collection {
	graph := dag.New[Pipeline]()
	graph.AddNode(0, New("default", 0))
	return &Collection{
		Graph:     graph,
		Providers: map[state.Argument]int64{},
		Root:      []int64{},
	}
}

// AdjNodesToPipelines converts a list of Nodes (with type Pipeline) to a list of Pipelines
func AdjNodesToPipelines(nodes []*dag.Node[Pipeline]) []Pipeline {
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

// SetProvider sets the provider of the argument 'arg' to the pipeline ID 'id'.
// If there is already a pipeline that provides the argument 'arg', then a 'pipeline.ErrorAmbiguousProvider' is returned.
// The provider for the argument is used to create an edge between the provider and pipelines that require the argument that is provided by it.
func (c *Collection) SetProvider(arg state.Argument, id int64) error {
	if _, ok := c.Providers[arg]; ok {
		return fmt.Errorf("ambiguous `Provides` for argument '%s (%s)'. Error: '%w'", arg.Key, arg.Type.String(), ErrorAmbiguousProvider)
	}
	c.Providers[arg] = id
	return nil
}

// BuidlEdges generates the edges in each pipeline based on the required and provided args of each step and pipeline.
func (c *Collection) BuildEdges(log logrus.FieldLogger, rootArgs ...state.Argument) error {
	c.Graph.Edges = map[int64][]dag.Edge[Pipeline]{}

	// Build the edges between each pipeline
	// Starting with pipelines that have no requirements
	for _, v := range c.Root {
		log.Debugln("Adding edge from root node to", v)
		if err := c.Graph.AddEdge(0, v); err != nil {
			return fmt.Errorf("error creating edge from '%d' to '%d': %w", 0, v, err)
		}
	}
	// Build the edges for the graph of steps in each pipeline
	for _, v := range c.Graph.Nodes {
		// default pipeline has no steps
		if v.ID == 0 {
			continue
		}
		// Find the node that provides the argument that we require.
		for _, arg := range v.Value.RequiredArgs {
			providerID, ok := c.Providers[arg]
			// If there is none, then we require an argument that nothing provides; do not continue.
			if !ok && arg.Type != state.ArgumentTypeSecret {
				return fmt.Errorf("%w: %s (%s)", ErrorNoPipelineProvider, arg.Key, arg.Type.String())
			}

			log.Debugln("Adding edge from provider (%d) to node (%d)", providerID, v.ID)
			// Add the edge to that node.
			if err := c.Graph.AddEdge(providerID, v.ID); err != nil {
				return err
			}
		}

		// Do pretty much the same thing for every step in our pipeline, too.
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

// pipelineVisitFunc returns a dag.VisitFunc that runs per-pipeline found in the graph.
func (c *Collection) pipelineVisitFunc(ctx context.Context, wf PipelineWalkFunc) dag.VisitFunc[Pipeline] {
	return func(n *dag.Node[Pipeline]) error {
		// Always skip the root node
		if n.ID == 0 {
			return nil
		}
		return wf(ctx, n.Value)
	}
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

// AppendPipeline adds a populated pipeline of Steps to the Graph.
func (c *Collection) AddPipelines(pipelines ...Pipeline) error {
	for _, v := range pipelines {
		if v.ID == 0 {
			continue
		}
		if err := c.Graph.AddNode(v.ID, v); err != nil {
			return fmt.Errorf("error adding pipeline node to graph: '%s', error: %w", v.Name, err)
		}
	}

	for _, v := range pipelines {
		id := v.ID
		if len(v.RequiredArgs) == 0 {
			// If this pipeline doesn't require anything before running, then it can run first.
			c.Root = append(c.Root, id)
		}

		for _, arg := range v.ProvidedArgs {
			if err := c.SetProvider(arg, id); err != nil {
				return fmt.Errorf("pipeline: '%s', error: %w", v.Name, err)
			}
		}
	}
	return nil
}

// ByID should return the Step that corresponds with a specific ID
func (c *Collection) ByID(ctx context.Context, id int64) (Step, error) {
	for _, p := range c.Graph.Nodes {
		for _, node := range p.Value.Graph.Nodes {
			if node.Value.ID == id {
				return node.Value, nil
			}
		}
	}

	return Step{}, errors.New("no step found")
}

// ByName should return the Step that corresponds with a specific Name
func (c *Collection) ByName(ctx context.Context, name string) ([]Step, error) {
	steps := []Step{}

	// Search every pipeline and step for the listed IDs
	if err := c.WalkPipelines(ctx, func(ctx context.Context, p Pipeline) error {
		if p.Name == name {
			return nil
		}
		for _, node := range p.Graph.Nodes {
			if node.Value.Name == name {
				steps = []Step{node.Value}
				return dag.ErrorBreak
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
	if err := c.WalkPipelines(ctx, func(ctx context.Context, p Pipeline) error {
		for i, argPipeline := range names {
			if strings.EqualFold(p.Name, argPipeline) {
				retP[i] = p
				found = true
				break
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
	for _, p := range c.Graph.Nodes {
		for _, event := range p.Value.Events {
			if event.Name == name {
				ret = append(ret, p.Value)
				break
			}
		}
	}

	if len(ret) == 0 {
		return nil, errors.New("no matching pipelines found")
	}

	return ret, nil
}
