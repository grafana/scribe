package shipwright

import (
	"fmt"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

type ShipwrightMulti struct {
	Client     pipeline.Client
	Collection *pipeline.Collection

	// Opts are the options that are provided to the pipeline from outside sources. This includes mostly command-line arguments and environment variables
	Opts pipeline.CommonOpts
	Log  logrus.FieldLogger

	n        *counter
	version  string
	pipeline int64

	prev []pipeline.Step[pipeline.Pipeline]
}

func (s *ShipwrightMulti) serial() int64 {
	return s.n.Next()
}

func (s *ShipwrightMulti) Run(step ...pipeline.Step[pipeline.Pipeline])      {}
func (s *ShipwrightMulti) Parallel(step ...pipeline.Step[pipeline.Pipeline]) {}
func (s *ShipwrightMulti) Sub(step ...pipeline.Step[pipeline.Pipeline])      {}
func (s *ShipwrightMulti) Done()                                             {}

// When allows users to define when this pipeline is executed, especially in the remote environment.
func (s *ShipwrightMulti) When(events ...pipeline.Event) {
	if err := s.Collection.AddEvents(s.pipeline, events...); err != nil {
		s.Log.WithError(err).Fatalln("Failed to add events to graph")
	}
}

// NewMulti is the equivalent of `shipwright.New`, but for building a pipeline made of multiple pipelines.
// Pipelines can behave in the same way that a step does. They can be ran in parallel using the Parallel function, or ran in a series using the Run function.
// To add new pipelines to execution, use the `(*shipwright[pipeline.Pipeline].New(...)` function.
func NewMulti() *Shipwright[pipeline.Pipeline] {
	opts, err := parseOpts()
	if err != nil {
		panic(fmt.Sprintf("failed to parse arguments: %s", err.Error()))
	}

	sw := NewClient[pipeline.Pipeline](opts, NewMultiCollection())

	// Ensure that no matter the behavior of the initializer, we still set the version on the shipwright object.
	sw.Version = opts.Args.Version

	return sw
}

type MultiFunc func(*Shipwright[pipeline.Action])

func MultiFuncWithLogging(logger logrus.FieldLogger, mf MultiFunc) MultiFunc {
	return func(sw *Shipwright[pipeline.Action]) {
		log := logger.WithFields(logrus.Fields{
			"n":        sw.n,
			"pipeline": sw.pipeline,
		})
		log.Debugln("Populating the sub pipeline...")
		mf(sw)
		log.Debugln("Done populating sub pipeline")
	}
}

// New creates a new Pipeline step that executes the provided MultiFunc onto a new `*Shipwright` type, creating a DAG.
// Because this function returns a pipeline.Step[T], it can be used with the normal Shipwright functions like `Run` and `Parallel`.
func (s *Shipwright[T]) New(name string, mf MultiFunc) pipeline.Step[pipeline.Pipeline] {
	log := s.Log.WithFields(logrus.Fields{
		"pipeline": name,
	})

	sw, err := s.newMulti(name)
	if err != nil {
		log.WithError(err).Fatalln("Failed to clone pipeline for use in multi-pipeline")
	}

	// This function adds the pipeline the way the user specified. It should look exactly like a normal shipwright pipeline.
	// This collection will be populated with a collection of Steps with actions.
	wrapped := MultiFuncWithLogging(log, mf)
	wrapped(sw)
	s.n = sw.n
	node, err := sw.Collection.Graph.Node(DefaultPipelineID)
	if err != nil {
		log.Fatal(err)
	}
	graph := node.Value.Content
	log.WithFields(logrus.Fields{
		"nodes": len(graph.Nodes),
		"edges": len(graph.Edges),
	}).Debugln("Graph populated")

	return pipeline.Step[pipeline.Pipeline]{
		Name:    name,
		Serial:  s.serial(),
		Content: node.Value.Content,
	}
}

func (s *Shipwright[T]) newMulti(name string) (*Shipwright[pipeline.Action], error) {
	log := s.Log.WithField("pipeline", name)
	collection := NewMultiCollection()

	if err := collection.AddPipelines(pipeline.NewPipelineNode(name, DefaultPipelineID)); err != nil {
		return nil, err
	}

	sw := &Shipwright[pipeline.Action]{
		Client:     s.Client,
		Opts:       s.Opts,
		Log:        log,
		Version:    s.Version,
		n:          s.n,
		Collection: collection,
		pipeline:   DefaultPipelineID,
	}

	return sw, nil
}
