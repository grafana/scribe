package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grafana/scribe/stringutil"
	swfs "github.com/grafana/scribe/swfs"
	"github.com/grafana/scribe/tarfs"
)

// FilesystemState stores state in a JSON file on the filesystem.
type FilesystemState struct {
	path string
	file string
	mtx  *sync.Mutex
}

func NewFilesystemState(path string) (*FilesystemState, error) {
	return &FilesystemState{
		path: path,
		mtx:  &sync.Mutex{},
	}, nil
}

// statePath gets the folder that can be used for placing items in the state.
// it should already exist.
func (f *FilesystemState) statePath() string {
	return filepath.Join(f.path, f.file)
}

// stateFile gets the path to the state.json file.
func (f *FilesystemState) stateFile() string {
	return filepath.Join(f.statePath(), "state.json")
}

func (f *FilesystemState) openr() (*os.File, error) {
	return os.Open(f.stateFile())
}

func (f *FilesystemState) openw() (*os.File, error) {
	return os.Create(f.stateFile())
}

func (f *FilesystemState) setValue(ctx context.Context, arg Argument, value any) error {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	r, err := f.openr()
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	state := JSONState{}

	// Error ignored intentionally. If there's an error, it's likely that the file is empty. We'll overwrite it.
	json.NewDecoder(r).Decode(&state)
	r.Close()

	// TODO: Do we really want to not allow overriding?
	// hmm
	// if _, ok := state[arg.Key]; ok {
	// 	return ErrorKeyExists
	// }

	w, err := f.openw()
	if err != nil {
		return err
	}

	defer w.Close()

	state[arg.Key] = StateValueJSON{
		Argument: arg,
		Value:    value,
	}

	return json.NewEncoder(w).Encode(state)
}

func (f *FilesystemState) getValue(ctx context.Context, arg Argument) (any, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	file, err := f.openr()
	if err != nil {
		return "", err
	}

	defer file.Close()

	state := map[string]StateValueJSON{}

	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return "", ErrorEmptyState
	}

	v, ok := state[arg.Key]
	if !ok {
		return "", ErrorNotFound
	}

	return v.Value, nil
}

func (f *FilesystemState) GetString(ctx context.Context, arg Argument) (string, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}

func (f *FilesystemState) SetString(ctx context.Context, arg Argument, value string) error {
	return f.setValue(ctx, arg, value)
}

func (f *FilesystemState) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return 0, err
	}

	return int64(v.(float64)), nil
}

func (f *FilesystemState) SetInt64(ctx context.Context, arg Argument, value int64) error {
	return f.setValue(ctx, arg, value)
}

func (f *FilesystemState) GetFloat64(ctx context.Context, arg Argument) (float64, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return 0, err
	}

	return v.(float64), nil
}

func (f *FilesystemState) SetFloat64(ctx context.Context, arg Argument, value float64) error {
	return f.setValue(ctx, arg, value)
}

func (f *FilesystemState) GetBool(ctx context.Context, arg Argument) (bool, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return false, err
	}

	return v.(bool), nil
}

func (f *FilesystemState) SetBool(ctx context.Context, arg Argument, value bool) error {
	return f.setValue(ctx, arg, value)
}

func (f *FilesystemState) GetFile(ctx context.Context, arg Argument) (*os.File, error) {
	v, err := f.getValue(ctx, arg)
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

func (f *FilesystemState) SetFile(ctx context.Context, arg Argument, value string) error {
	path := f.statePath()

	path = filepath.Join(path, filepath.Base(value))
	if err := swfs.CopyFile(value, path); err != nil {
		return err
	}

	return f.setValue(ctx, arg, path)
}

func (f *FilesystemState) SetFileReader(ctx context.Context, arg Argument, value io.Reader) (string, error) {
	path := f.statePath()

	path = filepath.Join(path, stringutil.Slugify(arg.Key))
	if err := swfs.CopyFileReader(value, path); err != nil {
		return "", err
	}

	return path, f.setValue(ctx, arg, path)
}

func (f *FilesystemState) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return nil, err
	}

	// Path will be the path to the tar.gz containing the directory, ending in `.tar.gz`.
	paths := v.(string)
	p := strings.Split(paths, ":")

	path := p[1]
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSuffix(path, filepath.Ext(path))
	fsp := f.statePath()

	destination := filepath.Join(fsp, name, stringutil.Random(8))

	// Extract the .tar.gz and provide the fs.FS
	// Ensure that this extraction is unique to this step.
	// TODO: maybe we can ensure that if multiple steps are using the same state directory, then we don't have to unzip it every time?
	if err := tarfs.Untar(destination, file); err != nil {
		return nil, err
	}

	return os.DirFS(destination), nil
}

// GetDirectoryString retrieves the original directory path.
// This can be particularly useful for things stored within the source filesystem.
func (f *FilesystemState) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	v, err := f.getValue(ctx, arg)
	if err != nil {
		return "", err
	}

	p := strings.Split(v.(string), ":")

	// This path will be the path to the directory and not the .tar.gz.
	return p[0], nil
}

func (f *FilesystemState) setDirectory(ctx context.Context, arg Argument, value string) error {
	// /tmp/asdf1234/x-asdf1234.tar.gz
	fsp := f.statePath()

	path := filepath.Join(fsp, fmt.Sprintf("%s-%s.tar.gz", stringutil.Slugify(arg.Key), stringutil.Random(8)))
	dir := os.DirFS(value)

	if _, err := tarfs.WriteFile(path, dir); err != nil {
		return fmt.Errorf("error creating tar.gz for directory state: %w", err)
	}

	return f.setValue(ctx, arg, strings.Join([]string{value, path}, ":"))
}

func (f *FilesystemState) setUnpackagedDirectory(ctx context.Context, arg Argument, value string) error {
	info, err := os.Stat(value)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("directory '%s' does not exist", value)
	}

	return f.setValue(ctx, arg, value)

}

func (f *FilesystemState) SetDirectory(ctx context.Context, arg Argument, value string) error {
	if arg.Type == ArgumentTypeFS {
		return f.setDirectory(ctx, arg, value)
	}
	return f.setUnpackagedDirectory(ctx, arg, value)
}

func (f *FilesystemState) Exists(ctx context.Context, arg Argument) (bool, error) {
	_, err := f.getValue(ctx, arg)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, ErrorNotFound) {
		return false, nil
	}

	return false, err
}
