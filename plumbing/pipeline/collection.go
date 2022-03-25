package pipeline

import "context"

// A StepGroup is a collection of Steps that are intended to run in parallel.
type StepGroup []Step

// WalkFunc is implemented by the executors. This function is executed for each step.
// If multiple steps are provided in the argument, then they were provided in "Parallel".
// If one step in the list of steps is of type "Background", then they all should be.
type WalkFunc func(context.Context, ...Step) error

// Walker is an interface that collections of steps should satisfy.
type Walker interface {
	Walk(context.Context, WalkFunc) error
}

// Collection defines a type that stores a collection of Steps.
type Collection interface {
	Walker

	// Append adds a new Step to the collection.
	Append(...Step) error

	// BySerial should return the Step that corresponds with a specific Serial
	BySerial(int) (Step, error)

	// ByName should return the Step that corresponds with a specific Name
	ByName(string) (Step, error)

	// Sub creates a new Collection of the same type from a list of Steps
	Sub(...Step) Collection
}
