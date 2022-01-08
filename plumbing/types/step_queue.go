package types

// A StepQueue is a queue of steps stored before a pipeline is ran.
type StepQueue struct {
	Steps []StepList
}

// At returns the step at the index 'v'. If no step is found, then nil is returned.
func (s *StepQueue) At(v int) StepList {
	if v >= len(s.Steps) {
		return nil
	}

	return s.Steps[v]
}

func (s *StepQueue) Append(steps ...Step) {
	s.Steps = append(s.Steps, StepList(steps))
}

// Next returns the next step in the queue.
// If there are no more steps, then nil is returned.
// This function removes the StepList from the queue
func (s *StepQueue) Next() StepList {
	if len(s.Steps) == 0 {
		return nil
	}

	step := s.Steps[0]

	s.Steps = s.Steps[1:]

	return step
}

func (s *StepQueue) Size() int {
	return len(s.Steps)
}
