package pipeline

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// FilesystemState looks in the filesystem for state values.
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

func (f *FilesystemState) file() (*os.File, error) {
	return os.OpenFile(f.path, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
}

func (f *FilesystemState) Get(arg Argument) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	file, err := f.file()
	if err != nil {
		return "", err
	}

	defer file.Close()

	state := map[string]string{}

	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return "", ErrorEmptyState
	}

	v, ok := state[arg.Key]
	if !ok {
		return "", ErrorNotFound
	}

	return v, nil
}

func (f *FilesystemState) Set(arg Argument, value io.Reader) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	file, err := f.file()
	if err != nil {
		return err
	}

	defer file.Close()

	state := map[string]StateValue{}

	if err := json.NewDecoder(file).Decode(&state); err != nil {
		// Do nothing, it's likely that the file is empty. We'll overwrite it.
	}

	if _, ok := state[arg.Key]; ok {
		return ErrorKeyExists
	}

	state[arg.Key] = StateValue{
		Argument: arg,
		Value:    value,
	}

	return json.NewEncoder(file).Encode(state)
}
