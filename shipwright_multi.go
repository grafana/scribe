package shipwright

import "github.com/grafana/shipwright/plumbing/pipeline"

func NewMulti() *Shipwright[pipeline.Pipeline] {
	return &Shipwright[pipeline.Pipeline]{}
}

type MultiFunc func(*Shipwright[pipeline.Action])

// Multi executes the provided MultiFunc onto a new `*Shipwright` type, creating a DAG.
// Because this function returns a pipeline.Step[T], it can be used with the normal Shipwright functions like `Run` and `Parallel`.
func (s *Shipwright[T]) Multi(name string, mf MultiFunc) pipeline.Step[pipeline.Pipeline] {
	sw := s.clone()

	// This function adds the pipeline the way the user specified. It should look exactly like a normal shipwright pipeline.
	// This collection will be populated with a collection of Steps with actions.
	mf(sw)

	return pipeline.Step[pipeline.Pipeline]{
		Serial: s.n,
	}
}

func (s *Shipwright[T]) clone() *Shipwright[pipeline.Action] {
	return &Shipwright[pipeline.Action]{
		Collection: NewMultiCollection(),
		Opts:       s.Opts,
		Log:        s.Log,
		Version:    s.Version,
		n:          s.n,
		pipeline:   s.pipeline + 1,
	}
}
