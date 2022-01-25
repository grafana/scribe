package types

type (
	StepAction func() error
	Argument   interface{}
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
	Arguments    []StepArgument

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

func (s Step) WithArguments(arg ...StepArgument) Step {
	s.Arguments = arg
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

// NoOpStep is used to represent a step which only exists to form uncommon relationships or for testing.
// Most clients should completely ignore NoOpSteps.
var NoOpStep = Step{
	Name: "no op",
	Action: func() error {
		return nil
	},
}
