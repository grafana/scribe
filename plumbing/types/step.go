package types

type StepAction func() error

// A Step stores a StepAction and a name for use in pipelines
type Step struct {
	Name   string
	Action StepAction
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
