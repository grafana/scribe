package pipeline

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	ErrorEmptyState = errors.New("state is empty")
	ErrorNotFound   = errors.New("key not found in state")
	ErrorKeyExists  = errors.New("key already exists in state")
	ErrorReadOnly   = errors.New("state is read-only")
)

type StateReader interface {
	GetString(Argument) (string, error)
	GetInt64(Argument) (int64, error)
	GetFloat64(Argument) (float64, error)
	GetFile(Argument) (*os.File, error)
	GetDirectory(Argument) (fs.FS, error)
}

type StateWriter interface {
	SetString(Argument, string) error
	SetInt64(Argument, int64) error
	SetFloat64(Argument, float64) error
	SetFile(Argument, string) error
	SetDirectory(Argument, string) error
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

// GetString attempts to get the string from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetString(arg Argument) (string, error) {
	if !ArgumentTypesEqual(ArgumentTypeString, arg) {
		return "", fmt.Errorf("attempted to get string from state for wrong argument type '%s'", arg.Type)
	}

	value, err := s.Handler.GetString(arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetString(arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return "", err
}

// GetInt64 attempts to get the int64 from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetInt64(arg Argument) (int64, error) {
	if !ArgumentTypesEqual(ArgumentTypeInt64, arg) {
		return 0, fmt.Errorf("attempted to get int64 from state for wrong argument type '%s'", arg.Type)
	}

	value, err := s.Handler.GetInt64(arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetInt64(arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return 0, err
}

// GetFloat64 attempts to get the int64 from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetFloat64(arg Argument) (float64, error) {
	if !ArgumentTypesEqual(ArgumentTypeFloat64, arg) {
		return 0.0, fmt.Errorf("attempted to get float64 from state for wrong argument type '%s'", arg.Type)
	}

	s.Log.Debugln("Getting float64 argument", arg.Key, "from state")
	value, err := s.Handler.GetFloat64(arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetFloat64(arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return 0, err
}

// GetFile attempts to get the file from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetFile(arg Argument) (*os.File, error) {
	if !ArgumentTypesEqual(ArgumentTypeFile, arg) {
		return nil, fmt.Errorf("attempted to get file from state for wrong argument type '%s'", arg.Type)
	}

	file, err := s.Handler.GetFile(arg)
	if err == nil {
		return file, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetFile(arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return nil, err
}

// GetDirectory attempts to get the directory from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetDirectory(arg Argument) (fs.FS, error) {
	if !ArgumentTypesEqual(ArgumentTypeFS, arg) {
		return nil, fmt.Errorf("attempted to get directory from state for wrong argument type '%s'", arg.Type)
	}

	dir, err := s.Handler.GetDirectory(arg)
	if err == nil {
		return dir, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		dir, err := v.GetDirectory(arg)
		if err == nil {
			return dir, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return nil, err
}

// SetString attempts to set the string into the state.
func (s *State) SetString(arg Argument, value string) error {
	if !ArgumentTypesEqual(ArgumentTypeString, arg) {
		return fmt.Errorf("attempted to set string in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetString(arg, value)
}

// SetInt64 attempts to set the int64 into the state.
func (s *State) SetInt64(arg Argument, value int64) error {
	if !ArgumentTypesEqual(ArgumentTypeInt64, arg) {
		return fmt.Errorf("attempted to set int64 in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetInt64(arg, value)
}

// SetFloat64 attempts to set the float64 into the state.
func (s *State) SetFloat64(arg Argument, value float64) error {
	if !ArgumentTypesEqual(ArgumentTypeFloat64, arg) {
		return fmt.Errorf("attempted to set float64 in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetFloat64(arg, value)
}

// SetFile attempts to set the file into the state.
// The "path" argument should be the path to the file to be stored.
func (s *State) SetFile(arg Argument, path string) error {
	if !ArgumentTypesEqual(ArgumentTypeFile, arg) {
		return fmt.Errorf("attempted to set file in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetFile(arg, path)
}

// SetDirectory attempts to set the directory into the state.
// The "path" argument should be the path to the directory to be stored.
func (s *State) SetDirectory(arg Argument, path string) error {
	if !ArgumentTypesEqual(ArgumentTypeFS, arg) {
		return fmt.Errorf("attempted to set folder in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetDirectory(arg, path)
}
