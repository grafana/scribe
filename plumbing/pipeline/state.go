package pipeline

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

var (
	ErrorEmptyState = errors.New("state is empty")
	ErrorNotFound   = errors.New("key not found in state")
	ErrorKeyExists  = errors.New("key already exists in state")
)

// State should be thread safe.
type State interface {
	Get(string) (string, error)
	Set(string, string) error
}

type FilesystemState struct {
	File *os.File
	mtx  *sync.Mutex
}

func NewFilesystemState(path string) (*FilesystemState, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.FileMode(0666))
	if err != nil {
		return nil, err
	}

	return &FilesystemState{
		File: f,
		mtx:  &sync.Mutex{},
	}, nil
}

func (f *FilesystemState) Get(key string) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	state := map[string]string{}

	if err := json.NewDecoder(f.File).Decode(&state); err != nil {
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

	state := map[string]string{}

	if err := json.NewDecoder(f.File).Decode(&state); err != nil {
		// Do nothing, it's likely that the file is empty. We'll overwrite it.
	}

	if _, ok := state[key]; ok {
		return ErrorKeyExists
	}

	state[key] = value

	return json.NewEncoder(f.File).Encode(state)
}
