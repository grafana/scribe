package state

import (
	"context"
	"io/fs"
	"os"
	"strconv"

	"github.com/grafana/scribe/args"
)

// ArgMapReader attempts to read state values from the provided "ArgMap".
// The ArgMap is provided by the user by using the '-arg={key}={value}' argument.
// This is typically only used in local executions where some values will not be provided.
type ArgMapReader struct {
	defaults args.ArgMap
}

func NewArgMapReader(defaults args.ArgMap) *ArgMapReader {
	return &ArgMapReader{
		defaults: defaults,
	}
}

func (s *ArgMapReader) GetString(ctx context.Context, arg Argument) (string, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *ArgMapReader) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(val, 10, 64)
}

func (s *ArgMapReader) GetFloat64(ctx context.Context, arg Argument) (float64, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return 0, err
	}

	return strconv.ParseFloat(val, 64)
}

func (s *ArgMapReader) GetBool(ctx context.Context, arg Argument) (bool, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return false, err
	}

	return strconv.ParseBool(val)
}

func (s *ArgMapReader) GetFile(ctx context.Context, arg Argument) (*os.File, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return nil, err
	}

	return os.Open(val)
}

func (s *ArgMapReader) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return nil, err
	}

	return os.DirFS(val), nil
}

func (s *ArgMapReader) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	val, err := s.defaults.Get(arg.Key)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *ArgMapReader) Exists(ctx context.Context, arg Argument) (bool, error) {
	// defaults.Get only returns an error if no value was found.
	_, err := s.defaults.Get(arg.Key)
	if err != nil {
		return false, nil
	}

	return true, nil
}
