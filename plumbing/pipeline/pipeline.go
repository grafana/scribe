package pipeline

import "github.com/grafana/shipwright/plumbing/pipeline/dag"

// A Pipeline is really similar to a Step, except that it contains a graph of steps rather than
// a single action. Just like a Step, it has dependencies, a name, an ID, etc.
type Pipeline struct {
	ID           int64
	Name         string
	Steps        *dag.Graph[StepList]
	Events       []Event
	Type         PipelineType
	Dependencies []Pipeline
}

// New creates a new Step that represents a pipeline.
func New(name string, id int64) Pipeline {
	return Pipeline{
		Name:  name,
		ID:    id,
		Steps: dag.New[StepList](),
	}
}

func PipelineNames(s []Pipeline) []string {
	v := make([]string, len(s))
	for i := range s {
		v[i] = s[i].Name
	}

	return v
}
