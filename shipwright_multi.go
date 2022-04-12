package shipwright

import "github.com/grafana/shipwright/plumbing/pipeline"

type ShipwrightMulti struct {
	Shipwright[pipeline.StepList]
}

func NewMulti() *ShipwrightMulti {
	return &ShipwrightMulti{}
}

// Multi executes the provided MultiFunc onto a new `*Shipwright` type, creating a DAG.
// Because this function returns a pipeline.Step[T], it can be used with the normal Shipwright functions like `Run` and `Parallel`.
func (s *ShipwrightMulti) Multi(name string, mf MultiFunc) pipeline.Step[pipeline.StepList] {
	sw := s.sub()

	// This function adds the pipeline the way the user specified. It should look exactly like a normal shipwright pipeline.
	// This collection will be populated with a collection of Steps with actions.
	mf(sw)

	return pipeline.Step[pipeline.StepList]{
		Serial: s.n,
	}
}

func (s *ShipwrightMulti) sub() Shipwright[pipeline.Action] {
	return Shipwright[pipeline.Action]{
		Collection: NewMultiCollection(),
		Opts:       s.Opts,
		Log:        s.Log,
		Version:    s.Version,
		n:          s.n,
		pipeline:   s.pipeline + 1,
	}
}
