package scribe

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/sirupsen/logrus"
)

type ScribeMulti struct {
	Client     pipeline.Client
	Collection *pipeline.Collection

	// Opts are the options that are provided to the pipeline from outside sources. This includes mostly command-line arguments and environment variables
	Opts    clients.CommonOpts
	Log     logrus.FieldLogger
	Version string

	n        *counter
	pipeline int64
}

func (s *ScribeMulti) serial() int64 {
	return s.n.Next()
}

// Add adds new pipelines to the Scribe DAG to be processed by the Client.
func (s *ScribeMulti) Add(pipelines ...pipeline.Pipeline) {
	for _, v := range pipelines {
		s.Log.WithFields(logrus.Fields{
			"name":     v.Name,
			"requires": v.RequiredArgs,
			"provides": v.ProvidedArgs,
		}).Debugln("adding pipeline")
	}
	if err := s.Collection.AddPipelines(pipelines...); err != nil {
		s.Log.WithError(err).Fatalln("error adding pipelines")
	}
}

// Execute is the equivalent of Done, but returns an error.
// Done should be preferred in Scribe pipelines as it includes sub-process handling and logging.
func (s *ScribeMulti) Execute(ctx context.Context, collection *pipeline.Collection) error {
	// Only worry about building an entire graph if we're not running a specific step.
	if step := s.Opts.Args.Step; step == nil || (*step) == 0 {
		rootArgs := pipeline.ClientProvidedArguments
		if err := collection.BuildEdges(s.Opts.Log, rootArgs...); err != nil {
			return err
		}
	}
	if err := s.Client.Done(ctx, collection); err != nil {
		return err
	}
	return nil
}

func (s *ScribeMulti) Done() {
	ctx := context.Background()
	if err := execute(ctx, s.Collection, nameOrDefault(s.Opts.Name), s.Opts, s.n, s.Execute); err != nil {
		s.Log.WithError(err).Fatal("error in execution")
	}
}

// NewMulti is the equivalent of `scribe.New`, but for building a pipeline made of multiple pipelines.
// Pipelines can behave in the same way that a step does. They can be ran in parallel using the Parallel function, or ran in a series using the Run function.
// To add new pipelines to execution, use the `(*scribe.ScribeMulti).New(...)` function.
func NewMulti() *ScribeMulti {
	rand.Seed(time.Now().Unix())
	ctx := context.Background()
	opts, err := parseOpts()
	if err != nil {
		panic(fmt.Sprintf("failed to parse arguments: %s", err.Error()))
	}

	sw := NewClient(ctx, opts, NewMultiCollection())

	return &ScribeMulti{
		Client:     sw.Client,
		Collection: sw.Collection,
		Opts:       opts,
		Log:        sw.Log,

		// Ensure that no matter the behavior of the initializer, we still set the version on the scribe object.
		Version: opts.Args.Version,
		n:       &counter{1},
	}
}

func NewMultiWithClient(opts clients.CommonOpts, client pipeline.Client) *ScribeMulti {
	rand.Seed(time.Now().Unix())
	if opts.Args == nil {
		opts.Args = &args.PipelineArgs{}
	}

	return &ScribeMulti{
		Client:     client,
		Opts:       opts,
		Log:        opts.Log,
		Collection: NewMultiCollection(),
		n:          &counter{1},
	}
}

type MultiFunc func(*Scribe)

func MultiFuncWithLogging(logger logrus.FieldLogger, mf MultiFunc) MultiFunc {
	return func(sw *Scribe) {
		log := logger.WithFields(logrus.Fields{
			"n":        sw.n,
			"pipeline": sw.pipeline,
		})
		log.Debugln("Populating the sub pipeline...")
		mf(sw)
		log.Debugln("Done populating sub pipeline")
	}
}

// New creates a new Pipeline step that executes the provided MultiFunc onto a new `*Scribe` type, creating a DAG.
// Because this function returns a pipeline.Step[T], it can be used with the normal Scribe functions like `Run` and `Parallel`.
func (s *ScribeMulti) New(name string, mf MultiFunc) pipeline.Pipeline {
	log := s.Log.WithFields(logrus.Fields{
		"pipeline": name,
	})
	sw, err := s.newMulti(name)
	if err != nil {
		log.WithError(err).Fatalln("Failed to clone pipeline for use in multi-pipeline")
	}

	sw.Opts.Name = name
	// This function adds the pipeline the way the user specified. It should look exactly like a normal scribe pipeline.
	// This collection will be populated with a collection of Steps with actions.
	wrappedMultiFunc := MultiFuncWithLogging(log, mf)
	wrappedMultiFunc(sw)

	// Update our counter with the new value of the sub-pipeline counter
	s.n = sw.n

	node, err := sw.Collection.Graph.Node(DefaultPipelineID)
	if err != nil {
		log.Fatal(err)
	}

	id := s.serial()
	log.WithFields(logrus.Fields{
		"nodes":    len(node.Value.Graph.Nodes),
		"requires": node.Value.RequiredArgs,
		"provides": node.Value.RequiredArgs,
		"name":     name,
		"id":       id,
	}).Debugln("Sub-pipeline created")

	return pipeline.Pipeline{
		ID:           id,
		Name:         name,
		Events:       node.Value.Events,
		Graph:        node.Value.Graph,
		Providers:    node.Value.Providers,
		Root:         node.Value.Root,
		RequiredArgs: node.Value.RequiredArgs,
		ProvidedArgs: node.Value.ProvidedArgs,
	}
}

func (s *ScribeMulti) newMulti(name string) (*Scribe, error) {
	log := s.Log.WithField("pipeline", name)
	collection := NewMultiCollection()
	if err := collection.AddPipelines(pipeline.New(name, DefaultPipelineID)); err != nil {
		return nil, err
	}

	sw := &Scribe{
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

func (s *ScribeMulti) PrintGraph(msg string) {
	for _, v := range s.Collection.Graph.Nodes {
		s.Log.WithFields(logrus.Fields{
			"id":       v.ID,
			"name":     v.Value.Name,
			"steps":    len(v.Value.Graph.Nodes),
			"edges":    len(v.Value.Graph.Edges),
			"requires": v.Value.RequiredArgs,
			"provides": v.Value.ProvidedArgs,
		}).Debugln(msg)
	}
}
