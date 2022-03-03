package pipeline

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// The ActionOpts are provided to every step that is ran.
// Each step can choose to use these options.
type ActionOpts struct {
	Stdout io.ReadWriter
	Stderr io.ReadWriter
}

type (
	StepAction func(context.Context, ActionOpts) error
	Output     interface{}
)

// A Step stores a StepAction and a name for use in pipelines
type Step struct {
	// Name is a string that represents or describes the step, essentially the identifier.
	// Not all run modes will support using the name.
	Name string

	// Image is an optional value that can be assigned to a step.
	// Typically, in docker environments (or drone with a Docker executor), it defines the docker image that is used to run the step.
	Image string

	// Action defines the action that this step takes in order to execute.
	Action StepAction

	// Dependencies define other steps that are required to run before this one.
	Dependencies []Step

	// Arguments are arguments that are must exist in order for this step to run.
	Arguments []Argument

	// Provides are arguments that this step provides for other arguments to use in their "Arguments" list.
	ProvidesArgs []Argument

	// Serial is the unique number that represents this step.
	// This value is used when calling `shipwright -step={serial} [pipeline]`
	Serial int
}

func (s Step) After(step Step) Step {
	if s.Dependencies == nil {
		s.Dependencies = []Step{}
	}

	s.Dependencies = append(s.Dependencies, step)

	return s
}

func (s Step) WithImage(image string) Step {
	s.Image = image
	return s
}

func (s Step) WithOutput(artifact Artifact) Step {
	return s
}

func (s Step) WithInput(artifact Artifact) Step {
	return s
}

func (s Step) WithArguments(arg ...Argument) Step {
	s.Arguments = arg
	return s
}

func (s Step) Provides(arg ...Argument) Step {
	s.ProvidesArgs = arg
	return s
}

func (s Step) WithName(name string) Step {
	s.Name = name
	return s
}

// NewStep creates a new step with an automatically generated name
func NewStep(action StepAction) Step {
	return Step{
		Action: action,
	}
}

// NamedStep creates a new step with an automatically generated name
func NamedStep(name string, action StepAction) Step {
	return Step{
		Name:   name,
		Action: action,
	}
}

// A StepList is a list of steps that are ran in parallel.
// This type is only used for intermittent storage and should not be used in the Shipwright client library
type StepList []Step

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

// NoOpStep is used to represent a step which only exists to form uncommon relationships or for testing.
// Most clients should completely ignore NoOpSteps.
var NoOpStep = Step{
	Name: "no op",
	Action: func(context.Context, ActionOpts) error {
		return nil
	},
}

// Combine combines the list of steps into one step, combining all of their required and provided arguments, as well as their actions.
// For string values that can not be combined, like Name and Image, the first step's values are chosen.
// These can be overridden with further chaining.
func Combine(step ...Step) Step {
	s := Step{
		Name:         step[0].Name,
		Image:        step[0].Image,
		Dependencies: []Step{},
		Arguments:    []Argument{},
		ProvidesArgs: []Argument{},
	}

	for _, v := range step {
		s.Dependencies = append(s.Dependencies, v.Dependencies...)
		s.Arguments = append(s.Arguments, v.Arguments...)
		s.ProvidesArgs = append(s.ProvidesArgs, v.ProvidesArgs...)
	}

	s.Action = func(ctx context.Context, opts ActionOpts) error {
		for _, v := range step {
			if err := v.Action(ctx, opts); err != nil {
				return err
			}
		}

		return nil
	}

	return s
}
