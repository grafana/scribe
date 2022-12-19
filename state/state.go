package state

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var (
	ErrorEmptyState = errors.New("state is empty")
	ErrorNotFound   = errors.New("key not found in state")
	ErrorKeyExists  = errors.New("key already exists in state")
	ErrorReadOnly   = errors.New("state is read-only")
)

type Reader interface {
	Exists(context.Context, Argument) (bool, error)
	GetString(context.Context, Argument) (string, error)
	GetInt64(context.Context, Argument) (int64, error)
	GetFloat64(context.Context, Argument) (float64, error)
	GetBool(context.Context, Argument) (bool, error)
	GetFile(context.Context, Argument) (*os.File, error)
	GetDirectory(context.Context, Argument) (fs.FS, error)
	GetDirectoryString(context.Context, Argument) (string, error)
}

type Writer interface {
	SetString(context.Context, Argument, string) error
	SetInt64(context.Context, Argument, int64) error
	SetFloat64(context.Context, Argument, float64) error
	SetBool(context.Context, Argument, bool) error
	SetFile(context.Context, Argument, string) error
	SetFileReader(context.Context, Argument, io.Reader) (string, error)
	SetDirectory(context.Context, Argument, string) error
}

type Handler interface {
	Reader
	Writer
}

type State struct {
	Handler  Handler
	Fallback []Reader
	Log      logrus.FieldLogger
}

// Exists checks the state to see if an argument exists in it.
// It can return an error in the event of a failure to check the state.
// An error will not be returned if the state could be read and the value was not in it.
// If a value for argument was not found, then false and a nil error is returned.
func (s *State) Exists(ctx context.Context, arg Argument) (bool, error) {
	exists, err := s.Handler.Exists(ctx, arg)
	if err != nil {
		return false, err
	}

	if exists {
		return true, nil
	}

	for _, v := range s.Fallback {
		exists, err := v.Exists(ctx, arg)
		if err != nil {
			return false, err
		}
		if exists {
			return exists, nil
		}
	}

	return false, nil
}

// GetString attempts to get the string from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetString(ctx context.Context, arg Argument) (string, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeString, ArgumentTypeSecret) {
		return "", fmt.Errorf("attempted to get string from state for wrong argument type '%s'", arg.Type)
	}

	value, err := s.Handler.GetString(ctx, arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetString(ctx, arg)
		if err == nil {
			if err := s.SetString(ctx, arg, val); err != nil {
				return "", err
			}
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return "", err
}

func MustGetString(s Handler, ctx context.Context, arg Argument) string {
	val, err := s.GetString(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// GetInt64 attempts to get the int64 from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeInt64) {
		return 0, fmt.Errorf("attempted to get int64 from state for wrong argument type '%s'", arg.Type)
	}

	value, err := s.Handler.GetInt64(ctx, arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetInt64(ctx, arg)
		if err == nil {
			if err := s.SetInt64(ctx, arg, value); err != nil {
				return 0, err
			}
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return 0, err
}

func MustGetInt64(s Handler, ctx context.Context, arg Argument) int64 {
	val, err := s.GetInt64(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// GetFloat64 attempts to get the int64 from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetFloat64(ctx context.Context, arg Argument) (float64, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeFloat64) {
		return 0.0, fmt.Errorf("attempted to get float64 from state for wrong argument type '%s'", arg.Type)
	}

	s.Log.Debugln("Getting float64 argument", arg.Key, "from state")
	value, err := s.Handler.GetFloat64(ctx, arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetFloat64(ctx, arg)
		if err == nil {
			if err := s.SetFloat64(ctx, arg, val); err != nil {
				return 0, err
			}
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return 0, err
}

func MustGetFloat64(s Handler, ctx context.Context, arg Argument) float64 {
	val, err := s.GetFloat64(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// GetBool attempts to get the bool from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetBool(ctx context.Context, arg Argument) (bool, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeBool) {
		return false, fmt.Errorf("attempted to get bool from state for wrong argument type '%s'", arg.Type)
	}

	s.Log.Debugln("Getting bool argument", arg.Key, "from state")
	value, err := s.Handler.GetBool(ctx, arg)
	if err == nil {
		return value, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetBool(ctx, arg)
		if err == nil {
			if err := s.SetBool(ctx, arg, val); err != nil {
				return false, err
			}
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return false, err
}

func MustGetBool(s Handler, ctx context.Context, arg Argument) bool {
	val, err := s.GetBool(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// GetFile attempts to get the file from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetFile(ctx context.Context, arg Argument) (*os.File, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeFile) {
		return nil, fmt.Errorf("attempted to get file from state for wrong argument type '%s'", arg.Type)
	}

	file, err := s.Handler.GetFile(ctx, arg)
	if err == nil {
		return file, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		val, err := v.GetFile(ctx, arg)
		if err == nil {
			return val, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return nil, err
}

func MustGetFile(s Handler, ctx context.Context, arg Argument) *os.File {
	val, err := s.GetFile(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// GetDirectory attempts to get the directory from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeFS, ArgumentTypeUnpackagedFS) {
		return nil, fmt.Errorf("attempted to get directory from state for wrong argument type '%s'", arg.Type)
	}

	dir, err := s.Handler.GetDirectoryString(ctx, arg)
	if err == nil {
		return os.DirFS(dir), nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		dir, err := v.GetDirectoryString(ctx, arg)
		if err == nil {
			dirAbs, err := filepath.Abs(dir)
			if err != nil {
				return nil, err
			}
			if err := s.SetDirectory(ctx, arg, dirAbs); err != nil {
				return nil, err
			}
			return os.DirFS(dir), nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return nil, err
}

// GetDirectory attempts to get the directory from the state.
// If there are Fallback readers and the state returned an error, then it will loop through each one, attempting to retrieve the value from the fallback state reader.
// If no fallback reader returns the value, then the original error is returned.
func (s *State) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeFS, ArgumentTypeUnpackagedFS) {
		return "", fmt.Errorf("attempted to get directory from state for wrong argument type '%s'", arg.Type)
	}

	dir, err := s.Handler.GetDirectoryString(ctx, arg)
	if err == nil {
		return dir, nil
	}

	for _, v := range s.Fallback {
		s.Log.WithError(err).Debugln("state returned an error; attempting fallback state")
		dir, err := v.GetDirectoryString(ctx, arg)
		if err == nil {
			dirAbs, err := filepath.Abs(dir)
			if err != nil {
				return "", fmt.Errorf("error getting absolute path from state value '%s': %w", dir, err)
			}
			if err := s.SetDirectory(ctx, arg, dirAbs); err != nil {
				return "", err
			}
			return dir, nil
		}

		s.Log.WithError(err).Debugln("fallback state reader returned an error")
	}

	return "", err
}

func MustGetDirectory(s Handler, ctx context.Context, arg Argument) fs.FS {
	val, err := s.GetDirectory(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

func MustGetDirectoryString(s Handler, ctx context.Context, arg Argument) string {
	val, err := s.GetDirectoryString(ctx, arg)
	if err != nil {
		panic(err)
	}

	return val
}

// SetString attempts to set the string into the state.
func (s *State) SetString(ctx context.Context, arg Argument, value string) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeString, ArgumentTypeSecret) {
		return fmt.Errorf("attempted to set string in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetString(ctx, arg, value)
}

// SetInt64 attempts to set the int64 into the state.
func (s *State) SetInt64(ctx context.Context, arg Argument, value int64) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeInt64) {
		return fmt.Errorf("attempted to set int64 in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetInt64(ctx, arg, value)
}

// SetFloat64 attempts to set the float64 into the state.
func (s *State) SetFloat64(ctx context.Context, arg Argument, value float64) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeFloat64) {
		return fmt.Errorf("attempted to set float64 in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetFloat64(ctx, arg, value)
}

// SetBool attempts to set the bool into the state.
func (s *State) SetBool(ctx context.Context, arg Argument, value bool) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeBool) {
		return fmt.Errorf("attempted to set bool in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetBool(ctx, arg, value)
}

// SetFile attempts to set the file into the state.
// The "path" argument should be the path to the file to be stored.
func (s *State) SetFile(ctx context.Context, arg Argument, path string) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeFile) {
		return fmt.Errorf("attempted to set file in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetFile(ctx, arg, path)
}

// SetFileReader attempts to set the reader into the state as a file.
// This is an easy way to go from downloading a file to setting it into the state without having to write it to disk first.
func (s *State) SetFileReader(ctx context.Context, arg Argument, r io.Reader) (string, error) {
	if !ArgumentTypesEqual(arg, ArgumentTypeFile) {
		return "", fmt.Errorf("attempted to set file in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetFileReader(ctx, arg, r)
}

// SetDirectory attempts to set the directory into the state.
// The "path" argument should be the path to the directory to be stored.
func (s *State) SetDirectory(ctx context.Context, arg Argument, path string) error {
	if !ArgumentTypesEqual(arg, ArgumentTypeUnpackagedFS, ArgumentTypeFS) {
		return fmt.Errorf("attempted to set folder in state for wrong argument type '%s'", arg.Type)
	}

	return s.Handler.SetDirectory(ctx, arg, path)
}
