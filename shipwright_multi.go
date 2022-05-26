package shipwright

import (
	"context"
	"fmt"

	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/sirupsen/logrus"
)

type ShipwrightMulti struct {
	Client     pipeline.Client
	Collection *pipeline.Collection

	// Opts are the options that are provided to the pipeline from outside sources. This includes mostly command-line arguments and environment variables
	Opts    pipeline.CommonOpts
	Log     logrus.FieldLogger
	Version string

	n        *counter
	pipeline int64

	prev []pipeline.Pipeline
}

func (s *ShipwrightMulti) serial() int64 {
	return s.n.Next()
}

// runPipeliens adds the list of pipelines to the collection. Pipelines are essentially branches in the graph.
// The pipelines provided run one after another.
func (s *ShipwrightMulti) runPipelines(pipelines ...pipeline.Pipeline) error {
	prev := s.prev

	for _, v := range pipelines {
		v.Dependencies = prev
		if err := s.Collection.AddPipelines(v); err != nil {
			return fmt.Errorf("error adding pipeline '%d' to collection. error: %w", v.ID, err)
		}

		prev = []pipeline.Pipeline{v}
	}

	s.prev = prev

	return nil
}

func (s *ShipwrightMulti) Run(steps ...pipeline.Pipeline) {
	if err := s.runPipelines(steps...); err != nil {
		s.Log.Fatalln(err)
	}
}

func (s *ShipwrightMulti) parallelPipelines(pipelines ...pipeline.Pipeline) error {
	for i := range pipelines {
		pipelines[i].Dependencies = s.prev
	}

	if err := s.Collection.AddPipelines(pipelines...); err != nil {
		return fmt.Errorf("error adding '%d' parallel pipelines to collection. error: %w", len(pipelines), err)
	}

	s.prev = pipelines

	return nil
}

func (s *ShipwrightMulti) Parallel(steps ...pipeline.Pipeline) {
	if err := s.parallelPipelines(steps...); err != nil {
		s.Log.Fatalln(err)
	}
}

func (s *ShipwrightMulti) subMulti(sub *ShipwrightMulti) error {
	prev := s.prev

	for i, v := range sub.Collection.Graph.Nodes {
		if v.ID == 0 || v.ID == DefaultPipelineID {
			continue
		}

		sub.Collection.Graph.Nodes[i].Value.Type = pipeline.PipelineTypeSub

		if len(v.Value.Dependencies) == 0 {
			sub.Collection.Graph.Nodes[i].Value.Dependencies = prev
		}

		if err := s.Collection.AddPipelines(sub.Collection.Graph.Nodes[i].Value); err != nil {
			return err
		}

		s.Log.Debugln("Appended pipeline", v.ID, v.Value.Name)
	}
	return nil
}

func (s *ShipwrightMulti) newSub() *ShipwrightMulti {
	serial := s.n.Next()
	opts := s.Opts
	opts.Name = fmt.Sprintf("sub-pipeline-%d", serial)

	collection := NewDefaultCollection(opts)

	return &ShipwrightMulti{
		Client:     s.Client,
		Opts:       opts,
		Log:        s.Log.WithField("sub-pipeline", opts.Name),
		Version:    s.Version,
		n:          s.n,
		Collection: collection,
		pipeline:   DefaultPipelineID,
	}
}

type MultiSubFunc func(*ShipwrightMulti)

func (s *ShipwrightMulti) Sub(sf MultiSubFunc) {
	sub := s.newSub()
	sf(sub)

	if err := s.subMulti(sub); err != nil {
		s.Log.WithError(err).Fatalln("failed to add sub-pipeline")
	}
}

// Execute is the equivalent of Done, but returns an error.
// Done should be preferred in Shipwright pipelines as it includes sub-process handling and logging.
func (s *ShipwrightMulti) Execute(ctx context.Context, collection *pipeline.Collection) error {
	if err := s.Client.Done(ctx, collection); err != nil {
		return err
	}
	return nil
}

func (s *ShipwrightMulti) Done() {
	ctx := context.Background()

	if err := execute(ctx, s.Collection, nameOrDefault(s.Opts.Name), s.Opts, s.n, s.Execute); err != nil {
		s.Log.WithError(err).Fatal("error in execution")
	}
}

// When allows users to define when this pipeline is executed, especially in the remote environment.
func (s *ShipwrightMulti) When(events ...pipeline.Event) {
	if err := s.Collection.AddEvents(s.pipeline, events...); err != nil {
		s.Log.WithError(err).Fatalln("Failed to add events to graph")
	}
}

// NewMulti is the equivalent of `shipwright.New`, but for building a pipeline made of multiple pipelines.
// Pipelines can behave in the same way that a step does. They can be ran in parallel using the Parallel function, or ran in a series using the Run function.
// To add new pipelines to execution, use the `(*shipwright[pipeline.Pipeline].New(...)` function.
func NewMulti() *ShipwrightMulti {
	opts, err := parseOpts()
	if err != nil {
		panic(fmt.Sprintf("failed to parse arguments: %s", err.Error()))
	}

	sw := NewClient(opts, NewMultiCollection())

	return &ShipwrightMulti{
		Client:     sw.Client,
		Collection: sw.Collection,
		Opts:       opts,
		Log:        sw.Log,

		// Ensure that no matter the behavior of the initializer, we still set the version on the shipwright object.
		Version: opts.Args.Version,
		n:       &counter{1},
	}
}

func NewMultiWithClient(opts pipeline.CommonOpts, client pipeline.Client) *ShipwrightMulti {
	if opts.Args == nil {
		opts.Args = &plumbing.PipelineArgs{}
	}

	return &ShipwrightMulti{
		Client:     client,
		Opts:       opts,
		Log:        opts.Log,
		Collection: NewMultiCollection(),
		n:          &counter{1},
	}
}

type MultiFunc func(*Shipwright)

func MultiFuncWithLogging(logger logrus.FieldLogger, mf MultiFunc) MultiFunc {
	return func(sw *Shipwright) {
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
func (s *ShipwrightMulti) New(name string, mf MultiFunc) pipeline.Pipeline {
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

	// Update our counter with the new value of the sub-pipeline counter
	s.n = sw.n

	node, err := sw.Collection.Graph.Node(DefaultPipelineID)
	if err != nil {
		log.Fatal(err)
	}
	graph := node.Value.Steps
	log.WithFields(logrus.Fields{
		"nodes": len(graph.Nodes),
		"edges": len(graph.Edges),
	}).Debugln("Graph populated")

	return pipeline.Pipeline{
		Name:  name,
		ID:    s.serial(),
		Steps: node.Value.Steps,
	}
}

func (s *ShipwrightMulti) newMulti(name string) (*Shipwright, error) {
	log := s.Log.WithField("pipeline", name)
	collection := NewMultiCollection()

	if err := collection.AddPipelines(pipeline.New(name, DefaultPipelineID)); err != nil {
		return nil, err
	}

	sw := &Shipwright{
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
