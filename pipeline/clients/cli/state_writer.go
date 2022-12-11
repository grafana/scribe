package cli

import (
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

func (w *StateWrapper) SetString(key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetString(key, val)
}

func (w *StateWrapper) SetInt64(key state.Argument, val int64) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetInt64(key, val)
}

func (w *StateWrapper) SetFloat64(key state.Argument, val float64) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetFloat64(key, val)
}

func (w *StateWrapper) SetBool(key state.Argument, val bool) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetBool(key, val)
}

func (w *StateWrapper) SetFile(key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetFile(key, val)
}

func (w *StateWrapper) SetFileReader(key state.Argument, r io.Reader) (string, error) {
	path, err := w.Writer.SetFileReader(key, r)
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    path,
	}
	return path, err
}

func (w *StateWrapper) SetDirectory(key state.Argument, val string) error {
	w.data[key.Key] = state.StateValueJSON{
		Argument: key,
		Value:    val,
	}
	return w.Writer.SetDirectory(key, val)
}

func (w *StateWrapper) Exists(arg state.Argument) (bool, error) {
	return w.Reader.Exists(arg)
}

func (w *StateWrapper) GetString(arg state.Argument) (string, error) {
	return w.Reader.GetString(arg)
}

func (w *StateWrapper) GetInt64(arg state.Argument) (int64, error) {
	return w.Reader.GetInt64(arg)
}

func (w *StateWrapper) GetFloat64(arg state.Argument) (float64, error) {
	return w.Reader.GetFloat64(arg)
}

func (w *StateWrapper) GetBool(arg state.Argument) (bool, error) {
	return w.Reader.GetBool(arg)
}

func (w *StateWrapper) GetFile(arg state.Argument) (*os.File, error) {
	return w.Reader.GetFile(arg)
}

func (w *StateWrapper) GetDirectory(arg state.Argument) (fs.FS, error) {
	return w.Reader.GetDirectory(arg)
}

func (w *StateWrapper) GetDirectoryString(arg state.Argument) (string, error) {
	return w.Reader.GetDirectoryString(arg)
}

func NewStateWrapper(r state.Reader, w state.Writer) *StateWrapper {
	return &StateWrapper{
		Reader: r,
		Writer: w,
		data:   make(map[string]state.StateValueJSON),
	}
}
