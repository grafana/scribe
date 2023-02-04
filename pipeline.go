package scribe

import (
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

// Pipeline is a more user-friendly, declarative representation of a Pipeline in the 'pipeline' package.
// This is only used when defining a pipeline in a declarative manner using the 'AddPipelines' function.
type Pipeline struct {
	Name     string
	Requires []state.Argument
	Steps    []pipeline.Step
	Provides []state.Argument
	When     []pipeline.Event
}

// AddPipelines adds a list of pipelines into the DAG. The order in which they are defined or added is not important; the order in which
// they run depends on what they require and what they provide.
// This function can be ran multiple times; every new item added with 'AddPipelines' will be appended to the dag.
func (s *ScribeMulti) AddPipelines(pipelines ...Pipeline) {
	for _, v := range pipelines {
		p := s.New(v.Name, func(s *Scribe) {
			if v.When != nil {
				s.When(v.When...)
			}
			s.Add(v.Steps...)
		})

		p = p.Requires(v.Requires...)
		p = p.Provides(v.Provides...)

		s.Add(p)
	}
}
