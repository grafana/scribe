package scribe

import (
	"context"
	"fmt"

	"github.com/grafana/scribe/v2/dag"
)

type ScribeOpts struct {
}

type ScribeOptFunc func(o *ScribeOpts)

// WithClient allows users to manually override the client for all pipeline runs.
func WithClient(c Client) ScribeOptFunc {
	return func(o *ScribeOpts) {
	}
}

// WithClientInitalizer allows users to add additional client initializers / options for the 'client' option.
func WithClientInitalizer(initializer ClientInitializerFunc) ScribeOptFunc {
	return func(o *ScribeOpts) {
	}
}

// FromFlags will load the appropriate options from CLI flags.
func FromFlags() ScribeOptFunc {
	return func(o *ScribeOpts) {
	}
}

// New returns a new Scribe client that matches the opts provided.
func New(opts ...ScribeOptFunc) *Scribe {
	ctx := context.Background()
	c, err := NewDaggerClient(ctx)
	if err != nil {
		panic(err)
	}

	return &Scribe{
		Pipelines: dag.New[Pipeline](),
		Client:    c,
		Providers: map[Argument]int64{},
		Requirers: map[Argument][]int64{},
	}
}

type Scribe struct {
	Client    Client
	Pipelines *dag.Graph[Pipeline]
	// Providers is a map of arguments and what node ID in the graph provides that argument.
	Providers map[Argument]int64
	// Requirers is a map of arguments and what node ID in the graph requires those argument.
	Requirers map[Argument][]int64

	i int64
}

func (s *Scribe) add(pipeline Pipeline) error {
	id := s.nextID()

	// Assign an ID to the pipeline
	s.Pipelines.AddNode(id, pipeline)

	// Label this pipeline as a provider for the arguments in 'Provides'.
	// If something already provides this argument, then throw an error. Multiple pipelines can not provide the same arguments.
	for _, v := range pipeline.Provides {
		if _, ok := s.Providers[v]; ok {
			return fmt.Errorf("argument '%s' is already provided by another pipeline", v.Key)
		}

		s.Providers[v] = id
	}

	for _, v := range pipeline.Requires {
		val, ok := s.Requirers[v]
		if !ok {
			val = []int64{}
		}

		s.Requirers[v] = append(val, id)
	}

	return nil
}

// Add adds the pipelines to the dag. All of the pipelines added to the dag are assigned an ID and are executed when "Run" is called.
func (s *Scribe) Add(pipelines ...Pipeline) error {
	for _, v := range pipelines {
		if err := s.add(v); err != nil {
			return err
		}
	}

	return nil
}

// Run will execute the dagger pipeline(s) that were added using the 'Add' function using the applicable client.
func (s *Scribe) Run() error {
	// Add graph edges based on requirers and providers.
	// Providers obviously need to run and set state before pipelines that require things do.
	// If a provider does not exist for argument then return an error.
	for reqArg, to := range s.Requirers {
		from, ok := s.Providers[reqArg]
		if !ok {
			return fmt.Errorf("argument '%s' is not provided by any pipeline", reqArg.Key)
		}
		for _, v := range to {
			// Create an edge from the 'provider' to each of the 'requirer'.
			if err := s.Pipelines.AddEdge(from, v); err != nil {
				return err
			}
		}
	}

	return s.Client.Run(s.Pipelines, ClientRunOpts{})
}

func (s *Scribe) nextID() int64 {
	s.i++

	return s.i
}
