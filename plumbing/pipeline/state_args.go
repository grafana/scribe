package pipeline

import (
	"errors"
	"log"

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
	log.Println("Getting key", key, "from arg map", s.defaults)
	log.Println("Getting key", key, "from arg map", s.defaults)
	log.Println("Getting key", key, "from arg map", s.defaults)
	log.Println("Getting key", key, "from arg map", s.defaults)

	val, err := s.defaults.Get(key)
	if err == nil {
		return val, nil
	}

	return "", errors.New("no defualt value found")
}
