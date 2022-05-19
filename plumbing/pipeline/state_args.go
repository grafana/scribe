package pipeline

import (
	"errors"

	"github.com/grafana/shipwright/plumbing"
)

type ArgMapReader struct {
	defaults plumbing.ArgMap
}

func NewArgMapReader(defaults plumbing.ArgMap) *ArgMapReader {
	return &ArgMapReader{
		defaults: defaults,
	}
}

func (s *ArgMapReader) Get(key string) (string, error) {
	val, err := s.defaults.Get(key)
	if err == nil {
		return val, nil
	}

	return "", errors.New("no defualt value found")
}
