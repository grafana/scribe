package shipwright

import (
	"fmt"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

// SubFunc should use the provided Shipwright object to populate a pipeline that runs independently.
type SubFunc[T pipeline.StepContent] func(*Shipwright[T])

// This function adds a single sub-pipeline
func (s *Shipwright[T]) subPipeline(sub *Shipwright[pipeline.Action]) error {
	node, err := sub.Collection.Graph.Node(DefaultPipelineID)
	if err != nil {
		return fmt.Errorf("Failed to retrieve populated subpipeline: %w", err)
	}

	p := node.Value
	p.Type = pipeline.StepTypeSubPipeline
	p.Serial = s.serial()
	if err := s.Collection.AddPipelines(p); err != nil {
		return err
	}

	return nil
}

func (s *Shipwright[T]) subPipelines(sub *Shipwright[pipeline.Pipeline]) error {
	prev := s.prevPipelines

	for i, v := range sub.Collection.Graph.Nodes {
		if v.ID == 0 || v.ID == DefaultPipelineID {
			continue
		}
		sub.Collection.Graph.Nodes[i].Value.Type = pipeline.StepTypeSubPipeline

		if len(v.Value.Dependencies) == 0 {
			sub.Collection.Graph.Nodes[i].Value.Dependencies = prev
		}

		if err := s.Collection.AddPipelines(sub.Collection.Graph.Nodes[i].Value); err != nil {
		}
		s.Log.Debugln("Appended pipeline", v.ID, v.Value.Name)
	}
	return nil
}

// Sub creates a sub-pipeline. The sub-pipeline is equivalent to creating a new coroutine made of multiple steps.
// This sub-pipeline will run concurrently with the rest of the pipeline at the time of definition.
// Under the hood, the Shipwright client creates a new Shipwright object with a clean Collection,
// then calles the SubFunc (sf) with the new Shipwright object. The collection is then populated by the SubFunc, and then appended to the existing collection.
func (s *Shipwright[T]) Sub(sf SubFunc[T]) {
	sub := s.newSub()

	s.Log.Debugf("Populating sub-pipeline in call to Sub")
	sf(sub)
	s.Log.Debugf("Populated sub-pipeline with '%d' nodes and '%d' edges", len(sub.Collection.Graph.Nodes), len(sub.Collection.Graph.Edges))
	s.Log.Debugf("Sub-pipeline nodes: '%+v'", pipeline.StepNames(pipeline.NodeListToSteps(sub.Collection.Graph.Nodes)))

	switch x := any(sub).(type) {
	case *Shipwright[pipeline.Action]:
		if err := s.subPipeline(x); err != nil {
			s.Log.WithError(err).Fatalln("failed to add sub-pipeline")
		}
	case *Shipwright[pipeline.Pipeline]:
		if err := s.subPipelines(x); err != nil {
			s.Log.WithError(err).Fatalln("failed to add sub-pipeline")
		}
	}
}

func (s *Shipwright[T]) newSub() *Shipwright[T] {
	serial := s.serial()
	opts := s.Opts
	opts.Name = fmt.Sprintf("sub-pipeline-%d", serial)

	collection := NewDefaultCollection(opts)

	return &Shipwright[T]{
		Client:     s.Client,
		Opts:       opts,
		Log:        s.Log.WithField("sub-pipeline", opts.Name),
		Version:    s.Version,
		n:          s.n,
		Collection: collection,
		pipeline:   DefaultPipelineID,
	}
}
