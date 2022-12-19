package cli

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/grafana/scribe/state"
)

type StateWrapper struct {
	state.Reader
	state.Writer
	data map[string]state.StateValueJSON
}

func (w *StateWrapper) SetString(ctx context.Context, key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetString(ctx, key, val)
}

func (w *StateWrapper) SetInt64(ctx context.Context, key state.Argument, val int64) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetInt64(ctx, key, val)
}

func (w *StateWrapper) SetFloat64(ctx context.Context, key state.Argument, val float64) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetFloat64(ctx, key, val)
}

func (w *StateWrapper) SetBool(ctx context.Context, key state.Argument, val bool) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetBool(ctx, key, val)
}

func (w *StateWrapper) SetFile(ctx context.Context, key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetFile(ctx, key, val)
}

func (w *StateWrapper) SetFileReader(ctx context.Context, key state.Argument, r io.Reader) (string, error) {
	path, err := w.Writer.SetFileReader(ctx, key, r)
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    path,
	}
	return path, err
}

func (w *StateWrapper) SetDirectory(ctx context.Context, key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetDirectory(ctx, key, val)
}

func (w *StateWrapper) Exists(ctx context.Context, arg state.Argument) (bool, error) {
	return w.Reader.Exists(ctx, arg)
}

func (w *StateWrapper) GetString(ctx context.Context, arg state.Argument) (string, error) {
	return w.Reader.GetString(ctx, arg)
}

func (w *StateWrapper) GetInt64(ctx context.Context, arg state.Argument) (int64, error) {
	return w.Reader.GetInt64(ctx, arg)
}

func (w *StateWrapper) GetFloat64(ctx context.Context, arg state.Argument) (float64, error) {
	return w.Reader.GetFloat64(ctx, arg)
}

func (w *StateWrapper) GetBool(ctx context.Context, arg state.Argument) (bool, error) {
	return w.Reader.GetBool(ctx, arg)
}

func (w *StateWrapper) GetFile(ctx context.Context, arg state.Argument) (*os.File, error) {
	return w.Reader.GetFile(ctx, arg)
}

func (w *StateWrapper) GetDirectory(ctx context.Context, arg state.Argument) (fs.FS, error) {
	return w.Reader.GetDirectory(ctx, arg)
}

func (w *StateWrapper) GetDirectoryString(ctx context.Context, arg state.Argument) (string, error) {
	return w.Reader.GetDirectoryString(ctx, arg)
}

func NewStateWrapper(r state.Reader, w state.Writer) *StateWrapper {
	return &StateWrapper{
		Reader: r,
		Writer: w,
		data:   make(map[string]state.StateValueJSON),
	}
}
