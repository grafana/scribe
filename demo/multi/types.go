package main

import (
	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

type Pipeline struct {
	Name     string
	Requires []state.Argument
	Steps    []pipeline.Step
	Provides []state.Argument
	When     []pipeline.Event
}

func AddPipelines(sw *scribe.ScribeMulti, pipelines []Pipeline) {
	for _, v := range pipelines {
		p := sw.New(v.Name, func(s *scribe.Scribe) {
			if v.When != nil {
				s.When(v.When...)
			}
			s.Add(v.Steps...)
		})

		p = p.Requires(v.Requires...)
		p = p.Provides(v.Provides...)

		sw.Add(p)
	}
}
