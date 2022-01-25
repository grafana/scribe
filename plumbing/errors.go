package plumbing

import (
	"errors"
	"fmt"
)

var (
	ErrorMissingArgument = errors.New("argument requested but not provided")
)

type PipelineError struct {
	Err         string
	Description string
}

func (p *PipelineError) Error() string {
	return fmt.Sprintf("%s: %s", p.Err, p.Description)
}

func NewPipelineError(err string, desc string) *PipelineError {
	return &PipelineError{
		Err:         err,
		Description: desc,
	}
}
