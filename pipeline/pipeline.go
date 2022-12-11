package pipeline

import (
	"errors"
	"fmt"

	"github.com/grafana/scribe/pipeline/dag"
	"github.com/grafana/scribe/state"
)

var (
	ErrorNoProvider        = errors.New("no step in the graph provides a required argument")
	ErrorAmbiguousProvider = errors.New("more than one step provides the same argument(s)")
)

// A Pipeline is really similar to a Step, except that it contains a graph of steps rather than
// a single action. Just like a Step, it has dependencies, a name, an ID, etc.
type Pipeline struct {
	ID   int64
	Name string
	// Graph is a graph where each node is a list of Steps. Each set of steps runs in parallel.
	Graph        *dag.Graph[Step]
	Providers    map[state.Argument]int64
	Root         []int64
	Events       []Event
	Type         PipelineType
	Dependencies []Pipeline
}

func (p *Pipeline) SetProvider(arg state.Argument, id int64) error {
	if _, ok := p.Providers[arg]; ok {
		return fmt.Errorf("ambiguous `Provides` for argument '%s (%s)'. Error: '%w'", arg.Key, arg.Type.String(), ErrorAmbiguousProvider)
	}
	p.Providers[arg] = id
	return nil
}

// AddStep adds all of the provided steps into the pipeline.
// The Step's ID field is used as the node ID.
func (p *Pipeline) AddSteps(steps ...Step) error {
	for _, v := range steps {
		id := v.ID
		// This node doesn't require anything before running, therefore it is a root node.
		if len(v.RequiredArgs) == 0 {
			p.Root = append(p.Root, id)
		}
		for _, arg := range v.ProvidedArgs {
			if err := p.SetProvider(arg, id); err != nil {
				return err
			}
		}
		if err := p.Graph.AddNode(id, v); err != nil {
			return err
		}
	}

	return nil
}

// BuildEdges generates the edges of the step graph based on the required / provided args of each step.
// It will return an error if there are required arguments that are not satisfied.
func (p *Pipeline) BuildEdges(rootArgs ...state.Argument) error {
	for _, v := range rootArgs {
		if err := p.SetProvider(v, 0); err != nil {
			return err
		}
	}
	// Clear the graph edges.
	p.Graph.Edges = map[int64][]dag.Edge[Step]{}
	// Every pipeline starts with a root node with an ID of 0.
	// Start by adding all of the 'p.Root' nodes to that node so that they run in parallel.
	for _, v := range p.Root {
		if err := p.Graph.AddEdge(0, v); err != nil {
			return err
		}
	}

	for _, node := range p.Graph.Nodes {
		for _, v := range node.Value.RequiredArgs {
			providerID, ok := p.Providers[v]
			if !ok && v.Type != state.ArgumentTypeSecret {
				return fmt.Errorf("%w: %s (%s)", ErrorNoProvider, v.Key, v.Type.String())
			}
			if err := p.Graph.AddEdge(providerID, node.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func nodeID(steps []Step) int64 {
	return steps[len(steps)-1].ID
}

// New creates a new Step that represents a pipeline.
func New(name string, id int64) Pipeline {
	graph := dag.New[Step]()
	graph.AddNode(0, Step{})
	return Pipeline{
		Name:  name,
		ID:    id,
		Graph: graph,
		Events: []Event{
			GitCommitEvent(GitCommitFilters{}),
		},
		Providers: map[state.Argument]int64{},
		Root:      []int64{},
	}
}

func PipelineNames(s []Pipeline) []string {
	v := make([]string, len(s))
	for i := range s {
		v[i] = s[i].Name
	}

	return v
}
