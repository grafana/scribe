package pipeline

import (
	"encoding/json"
	"os"
	"sync"
)

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

func (f *FilesystemState) Get(key string) (string, error) {
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

	v, ok := state[key]
	if !ok {
		return "", ErrorNotFound
	}

	return v, nil
}

func (f *FilesystemState) Set(key, value string) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	file, err := f.file()
	if err != nil {
		return err
	}

	defer file.Close()

	state := map[string]string{}

	if err := json.NewDecoder(file).Decode(&state); err != nil {
		// Do nothing, it's likely that the file is empty. We'll overwrite it.
	}

	if _, ok := state[key]; ok {
		return ErrorKeyExists
	}

	state[key] = value

	return json.NewEncoder(file).Encode(state)
}
