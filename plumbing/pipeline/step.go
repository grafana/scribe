package pipeline

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/grafana/shipwright/plumbing/pipeline/dag"
	"github.com/opentracing/opentracing-go"
)

// The ActionOpts are provided to every step that is ran.
// Each step can choose to use these options.
type ActionOpts struct {
	Stdout io.Writer
	Stderr io.Writer
	Tracer opentracing.Tracer
}

type (
	StepType int

	// Action is the function signature that a step provides when it does something.
	Action func(context.Context, ActionOpts) error
	Output interface{}

	// A StepList is a list of steps that are ran in parallel.
	// This type is only used for intermittent storage and should not be used in the Shipwright client library
	StepList []Step[Action]
	Pipeline struct{ *dag.Graph[Step[StepList]] }
)

const (
	StepTypeDefault StepType = iota
	StepTypeBackground
	StepTypeList
)

// StepContent is used as a type argument to the "Step" type.
// * Step[Action] is a Step that performs a single action. This type mostly exists for use by pipeline developers to define a single step that performs a single action.
// * Step[StepList] is a Step that stores multiple Steps that have actions. This is used for storage purposes in the internal data DAG structure.
// * Step[Pipeline] is a Step that stores
type StepContent interface {
	Action | StepList | Pipeline
}

// A Step stores a Action and a name for use in pipelines.
// A Step can consist of either a single action or represent a list of actions.
type Step[T StepContent] struct {
	// Type represents the how the step is intended to operate. 90% of the time, the default type should be a sufficient descriptor of a step.
	// However in some circumstances, clients may want to handle a step differently based on how it's defined.
	// Background steps, for example, have to have their lifecycles handled differently.
	Type StepType

	// Name is a string that represents or describes the step, essentially the identifier.
	// Not all run modes will support using the name.
	Name string

	// Image is an optional value that can be assigned to a step.
	// Typically, in docker environments (or drone with a Docker executor), it defines the docker image that is used to run the step.
	Image string

	// Content defines the contents of this step
	Content T

	// Dependencies define other steps that are required to run before this one.
	// As far as we're concerned, Steps can only depend on other steps of the same type.
	Dependencies []Step[T]

	// Arguments are arguments that are must exist in order for this step to run.
	Arguments []Argument

	// Provides are arguments that this step provides for other arguments to use in their "Arguments" list.
	ProvidesArgs []Argument

	// Serial is the unique number that represents this step.
	// This value is used when calling `shipwright -step={serial} [pipeline]`
	Serial int64
}

func (s Step[T]) IsBackground() bool {
	return s.Type == StepTypeBackground
}

func (s Step[T]) After(step Step[T]) Step[T] {
	if s.Dependencies == nil {
		s.Dependencies = []Step[T]{}
	}

	s.Dependencies = append(s.Dependencies, step)

	return s
}

func (s Step[T]) WithImage(image string) Step[T] {
	s.Image = image
	return s
}

func (s Step[T]) WithOutput(artifact Artifact) Step[T] {
	return s
}

func (s Step[T]) WithInput(artifact Artifact) Step[T] {
	return s
}

func (s Step[T]) WithArguments(arg ...Argument) Step[T] {
	s.Arguments = arg
	return s
}

func (s Step[T]) Provides(arg ...Argument) Step[T] {
	s.ProvidesArgs = arg
	return s
}

func (s Step[T]) WithName(name string) Step[T] {
	s.Name = name
	return s
}

// NewStep creates a new step with an automatically generated name
func NewStep(action Action) Step[Action] {
	return Step[Action]{
		Content: action,
	}
}

// NamedStep creates a new step with a name provided
func NamedStep(name string, action Action) Step[Action] {
	return Step[Action]{
		Name:    name,
		Content: action,
	}
}

func (s *StepList) Names() []string {
	names := make([]string, len(*s))

	for i, v := range *s {
		names[i] = v.Name
	}

	return names
}

func (s *StepList) String() string {
	return fmt.Sprintf("[%s]", strings.Join(s.Names(), " | "))
}

// DefaultAction is a nil action intentionally. In some client implementations, a nil step indicates a specific behavior.
// In Drone and Docker, for example, a nil step indicates that the docker command or entrypoint should not be supplied, thus using the default command for that image.
var DefaultAction Action = nil

// NoOpStep is used to represent a step which only exists to form uncommon relationships or for testing.
// Most clients should completely ignore NoOpSteps.
var NoOpStep = Step[Action]{
	Name: "no op",
	Content: func(context.Context, ActionOpts) error {
		return nil
	},
}

// Combine combines the list of steps into one step, combining all of their required and provided arguments, as well as their actions.
// For string values that can not be combined, like Name and Image, the first step's values are chosen.
// These can be overridden with further chaining.
func Combine(step ...Step[Action]) Step[Action] {
	s := Step[Action]{
		Name:         step[0].Name,
		Image:        step[0].Image,
		Dependencies: []Step[Action]{},
		Arguments:    []Argument{},
		ProvidesArgs: []Argument{},
	}

	for _, v := range step {
		s.Dependencies = append(s.Dependencies, v.Dependencies...)
		s.Arguments = append(s.Arguments, v.Arguments...)
		s.ProvidesArgs = append(s.ProvidesArgs, v.ProvidesArgs...)
	}

	s.Content = func(ctx context.Context, opts ActionOpts) error {
		for _, v := range step {
			if err := v.Content(ctx, opts); err != nil {
				return err
			}
		}

		return nil
	}

	return s
}
