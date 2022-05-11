package pipeline

import (
	"errors"

	"github.com/sirupsen/logrus"
)

var (
	ErrorEmptyState = errors.New("state is empty")
	ErrorNotFound   = errors.New("key not found in state")
	ErrorKeyExists  = errors.New("key already exists in state")
	ErrorReadOnly   = errors.New("state is read-only")
)

type StateReader interface {
	Get(string) (string, error)
}

type StateWriter interface {
	Set(string, string) error
}

type StateHandler interface {
	StateReader
	StateWriter
}

type State struct {
	Handler  StateHandler
	Fallback []StateReader
	Log      logrus.FieldLogger
}

func (s *State) Get(key string) (string, error) {
	value, err := s.Handler.Get(key)
	if err == nil {
		return value, nil
	}
	for _, v := range s.Fallback {
		val, err := v.Get(key)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return "", err
}

func (s *State) Set(key, value string) error {
	s.Log.Infof("Set key: %s, value: %s in state", key, value)
	s.Log.Infof("Set key: %s, value: %s in state", key, value)
	s.Log.Infof("Set key: %s, value: %s in state", key, value)
	s.Log.Infof("Set key: %s, value: %s in state", key, value)
	return s.Handler.Set(key, value)
}
