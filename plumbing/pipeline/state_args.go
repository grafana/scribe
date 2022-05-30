package pipeline

import (
	"errors"

	"github.com/grafana/shipwright/plumbing"
)

// ArgMapReader attempts to read state values from the provided "ArgMap".
// The ArgMap is provided by the user by using the '-arg={key}={value}' argument.
// This is typically only used in local executions where some values will not be provided.
type ArgMapReader struct {
	defaults plumbing.ArgMap
}

func NewArgMapReader(defaults plumbing.ArgMap) *ArgMapReader {
	return &ArgMapReader{
		defaults: defaults,
	}
}

func (s *ArgMapReader) Get(arg Argument) (StateValue, error) {
	val, err := s.defaults.Get(arg.Key)
	if err == nil {
		return StateValue{
			Argument: arg,
			Value:    val,
		}, nil
	}

	return "", errors.New("no defualt value found")
}
