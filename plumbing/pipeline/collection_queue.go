package pipeline

import (
	"context"
	"fmt"
)

// A Queue is a queue of steps stored before a pipeline is ran.
type Queue struct {
	Steps []StepGroup
}

// Append appends a new list of Steps to the queue.
// All of the steps provided to a single function call are intended to be ran in parallel.
func (s *Queue) Append(steps ...Step) error {
	s.Steps = append(s.Steps, StepGroup(steps))

	return nil
}

// Next returns the next step in the queue.
// If there are no more steps, then nil is returned.
// This function removes the StepGroup from the queue.
func (s *Queue) Next() StepGroup {
	if len(s.Steps) == 0 {
		return nil
	}

	step := s.Steps[0]
	s.Steps = s.Steps[1:]

	return step
}

// BySerial should return the Step that corresponds with a specific serial number.
func (s *Queue) BySerial(serial int) (Step, error) {
	for _, groups := range s.Steps {
		for _, v := range groups {
			if v.Serial == serial {
				return v, nil
			}
		}
	}

	return Step{}, fmt.Errorf("error: %w, id: %d", ErrorStepNotFound, serial)
}

// ByName should return the Step that corresponds with a specific name.
func (s *Queue) ByName(name string) (Step, error) {
	for _, groups := range s.Steps {
		for _, v := range groups {
			if v.Name == name {
				return v, nil
			}
		}
	}

	return Step{}, fmt.Errorf("error: %w, name: %s", ErrorStepNotFound, name)
}

// Walk executes the provided WalkFunc (wf) several times, providing parallel steps in the order they should be executed until none are left.
func (s *Queue) Walk(ctx context.Context, wf WalkFunc) error {
	for v := s.Next(); v != nil; v = s.Next() {
		if err := wf(ctx, v...); err != nil {
			return err
		}
	}

	return nil
}

func (s *Queue) Size() int {
	return len(s.Steps)
}

func (s *Queue) Sub(step ...Step) Collection {
	queue := NewQueue()
	for _, v := range step {
		queue.Append(v)
	}

	return queue
}

func NewQueue() *Queue {
	return &Queue{
		Steps: []StepGroup{},
	}
}
