package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	swfs "github.com/grafana/shipwright/fs/x"
	"github.com/grafana/shipwright/plumbing/stringutil"
	"github.com/grafana/shipwright/plumbing/tarfs"
)

type stateValue struct {
	Argument Argument `json:"argument"`
	Value    any      `json:"value"`
}

// FilesystemState stores state in a JSON file on the filesystem.
type FilesystemState struct {
	path string
	mtx  *sync.Mutex
}

func NewFilesystemState(path string) (*FilesystemState, error) {
	return &FilesystemState{
		path: path,
		mtx:  &sync.Mutex{},
	}, nil
}

func (f *FilesystemState) fsStatePath() string {
	return strings.TrimSuffix(f.path, filepath.Ext(f.path))
}

func (f *FilesystemState) openr() (*os.File, error) {
	return os.Open(f.path)
}

func (f *FilesystemState) openw() (*os.File, error) {
	return os.Create(f.path)
}

func (f *FilesystemState) setValue(arg Argument, value any) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	r, err := f.openr()
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	state := map[string]stateValue{}

	if err := json.NewDecoder(r).Decode(&state); err != nil {
		// Do nothing, it's likely that the file is empty. We'll overwrite it.
	}
	r.Close()

	if _, ok := state[arg.Key]; ok {
		return ErrorKeyExists
	}

	w, err := f.openw()
	if err != nil {
		return err
	}

	defer w.Close()

	state[arg.Key] = stateValue{
		Argument: arg,
		Value:    value,
	}

	return json.NewEncoder(w).Encode(state)
}

func (f *FilesystemState) getValue(arg Argument) (any, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	file, err := f.openr()
	if err != nil {
		return "", err
	}

	defer file.Close()

	state := map[string]stateValue{}

	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return "", ErrorEmptyState
	}

	v, ok := state[arg.Key]
	if !ok {
		return "", ErrorNotFound
	}

	return v.Value, nil

}

func (f *FilesystemState) GetString(arg Argument) (string, error) {
	v, err := f.getValue(arg)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}

func (f *FilesystemState) SetString(arg Argument, value string) error {
	return f.setValue(arg, value)
}

func (f *FilesystemState) GetInt64(arg Argument) (int64, error) {
	v, err := f.getValue(arg)
	if err != nil {
		return 0, err
	}

	return int64(v.(float64)), nil
}

func (f *FilesystemState) SetInt64(arg Argument, value int64) error {
	return f.setValue(arg, value)
}

func (f *FilesystemState) GetFloat64(arg Argument) (float64, error) {
	v, err := f.getValue(arg)
	if err != nil {
		return 0, err
	}

	return v.(float64), nil
}

func (f *FilesystemState) SetFloat64(arg Argument, value float64) error {
	return f.setValue(arg, value)
}

func (f *FilesystemState) GetFile(arg Argument) (*os.File, error) {
	v, err := f.getValue(arg)
	if err != nil {
		return nil, err
	}

	path := v.(string)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (f *FilesystemState) SetFile(arg Argument, value string) error {
	path := f.fsStatePath()
	path = filepath.Join(path, filepath.Base(value))
	if err := swfs.CopyFile(value, path); err != nil {
		return err
	}

	return f.setValue(arg, path)
}

func (f *FilesystemState) GetDirectory(arg Argument) (fs.FS, error) {
	v, err := f.getValue(arg)
	if err != nil {
		return nil, err
	}

	// Path will be the path to the tar.gz containing the directory, ending in `.tar.gz`.
	path := v.(string)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSuffix(path, filepath.Ext(path))
	destination := filepath.Join(f.fsStatePath(), name, stringutil.Random(8))

	// Extract the .tar.gz and provide the fs.FS
	// Ensure that this extraction is unique to this step.
	// TODO: maybe we can ensure that if multiple steps are using the same state directory, then we don't have to unzip it every time?
	if err := tarfs.Untar(destination, file); err != nil {
		return nil, err
	}

	return os.DirFS(destination), nil
}

func (f *FilesystemState) SetDirectory(arg Argument, value string) error {
	// /tmp/asdf1234/x-asdf1234.tar.gz
	path := filepath.Join(f.fsStatePath(), fmt.Sprintf("%s-%s.tar.gz", stringutil.Slugify(arg.Key), stringutil.Random(8)))
	dir := os.DirFS(value)

	_, err := tarfs.WriteFile(path, dir)
	if err != nil {
		return fmt.Errorf("error creating tar.gz for directory state: %w", err)
	}

	return f.setValue(arg, path)
}
