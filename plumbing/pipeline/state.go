package pipeline

import (
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

var (
	ErrorEmptyState = errors.New("state is empty")
	ErrorNotFound   = errors.New("key not found in state")
	ErrorKeyExists  = errors.New("key already exists in state")
	ErrorReadOnly   = errors.New("state is read-only")
)

type StateValue struct {
	Argument Argument `json:"argument"`

	// Value is a []byte in this case because it comes from an io.Reader. In order to support more types of encoding, I figured a []byte is an apt type for storing most relevant data in JSON.
	Value []byte `json:"value"`
}

func (s StateValue) String() string {
	return string(s.Value)
}

type StateReader interface {
	Get(Argument) (StateValue, error)
}

type StateWriter interface {
	Set(Argument, io.Reader) error
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

func (s *State) Get(arg Argument) (StateValue, error) {
	value, err := s.Handler.Get(arg)
	if err == nil {
		return value, nil
	}
	for _, v := range s.Fallback {
		val, err := v.Get(arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return StateValue{}, err
}

func (s *State) Set(key Argument, value io.Reader) error {
	return s.Handler.Set(key, value)
}
