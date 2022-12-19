package state

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/grafana/scribe/stringutil"
	"github.com/grafana/scribe/tarfs"
)

var (
	ErrorFileNotFound = errors.New("not found")
)

type ObjectStorageHandler struct {
	Storage  ObjectStorage
	Bucket   string
	BasePath string
	mtx      *sync.Mutex
}

func NewObjectStorageHandler(storage ObjectStorage, bucket, base string) *ObjectStorageHandler {
	return &ObjectStorageHandler{
		Storage:  storage,
		Bucket:   bucket,
		BasePath: base,
		mtx:      &sync.Mutex{},
	}
}

func (s *ObjectStorageHandler) stateKey(arg Argument) string {
	suffix := stringutil.Slugify(arg.Key)
	return path.Join(s.BasePath, "state", fmt.Sprintf("%s.json", suffix))
}

// readStateFile opens the state file at {bucketPath}/{basePath}/state.json and parses it.
func (s *ObjectStorageHandler) readStateFile(ctx context.Context, arg Argument) (JSONState, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	res, err := s.Storage.GetObject(ctx, s.Bucket, s.stateKey(arg))
	if err != nil {
		if errors.Is(err, ErrorFileNotFound) {
			return JSONState{}, nil
		}

		return nil, err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Println("error closing response body", err)
		}
	}()
	st := JSONState{}
	if err := json.NewDecoder(res.Body).Decode(&st); err == nil {
		return st, nil
	}

	return nil, nil
}

func (s *ObjectStorageHandler) getValue(ctx context.Context, arg Argument) (any, error) {
	st, err := s.readStateFile(ctx, arg)
	if err != nil {
		return nil, err
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	v, ok := st[arg.Key]
	if !ok {
		return "", ErrorNotFound
	}

	return v.Value, nil
}

// setValue opens and reads the state, updates it, and then re-uploads it.
// perhaps not the most efficient thing in the world but state reads and changes shouldn't happen very often.
func (s *ObjectStorageHandler) setValue(ctx context.Context, arg Argument, value any) error {
	st, err := s.readStateFile(ctx, arg)
	if err != nil {
		return err
	}
	s.mtx.Lock()
	defer s.mtx.Unlock()

	st[arg.Key] = StateValueJSON{
		Argument: arg,
		Value:    value,
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(st); err != nil {
		return err
	}

	if err := s.Storage.PutObject(ctx, s.Bucket, s.stateKey(arg), buf); err != nil {
		return err
	}

	return nil
}

func (s *ObjectStorageHandler) Exists(ctx context.Context, arg Argument) (bool, error) {
	_, err := s.getValue(ctx, arg)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, ErrorNotFound) {
		return false, nil
	}

	return false, err
}

func (s *ObjectStorageHandler) GetString(ctx context.Context, arg Argument) (string, error) {
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return "", err
	}

	return v.(string), nil
}

func (s *ObjectStorageHandler) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return 0, err
	}

	return int64(v.(float64)), nil
}

func (s *ObjectStorageHandler) GetFloat64(ctx context.Context, arg Argument) (float64, error) {
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return 0, err
	}

	return v.(float64), nil

}

func (s *ObjectStorageHandler) GetBool(ctx context.Context, arg Argument) (bool, error) {
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return false, err
	}

	return v.(bool), nil

}

func (s *ObjectStorageHandler) GetFile(ctx context.Context, arg Argument) (*os.File, error) {
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return nil, err
	}

	res, err := s.Storage.GetObject(ctx, s.Bucket, v.(string))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	dir := filepath.Join(os.TempDir(), stringutil.Random(8))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("error creating directory '%s' for file from object storage: %w", dir, err)
	}

	fileName := stringutil.Slugify(arg.Key)
	path := filepath.Join(dir, fileName)

	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("error writing file from object storage to filesystem: %w", err)
	}

	defer f.Close()
	io.Copy(f, res.Body)

	return f, nil
}

func (s *ObjectStorageHandler) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) {
	str, err := s.GetDirectoryString(ctx, arg)
	if err != nil {
		return nil, err
	}

	return os.DirFS(str), nil
}

func (s *ObjectStorageHandler) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	// Download the tarball, provide it as an fs.FS
	v, err := s.getValue(ctx, arg)
	if err != nil {
		return "", err
	}
	res, err := s.Storage.GetObject(ctx, s.Bucket, v.(string))
	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	path := filepath.Join(os.TempDir(), stringutil.Slugify(arg.Key))
	if err := tarfs.Untar(path, res.Body); err != nil {
		return "", err
	}

	return path, nil
}

func (s *ObjectStorageHandler) SetString(ctx context.Context, arg Argument, value string) error {
	return s.setValue(ctx, arg, value)
}

func (s *ObjectStorageHandler) SetInt64(ctx context.Context, arg Argument, value int64) error {
	return s.setValue(ctx, arg, value)
}

func (s *ObjectStorageHandler) SetFloat64(ctx context.Context, arg Argument, value float64) error {
	return s.setValue(ctx, arg, value)
}

func (s *ObjectStorageHandler) SetBool(ctx context.Context, arg Argument, value bool) error {
	return s.setValue(ctx, arg, value)
}

func (s *ObjectStorageHandler) SetFile(ctx context.Context, arg Argument, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	if _, err := s.SetFileReader(ctx, arg, file); err != nil {
		return err
	}

	return nil
}

func (s *ObjectStorageHandler) SetFileReader(ctx context.Context, arg Argument, r io.Reader) (string, error) {
	key := path.Join(s.BasePath, stringutil.Slugify(arg.Key))

	if err := s.Storage.PutObject(ctx, s.Bucket, key, r); err != nil {
		return "", err
	}

	return key, s.setValue(ctx, arg, key)
}

// setDirectory packages the provided directory and uploads it to the state as a packaged tar.gz
func (s *ObjectStorageHandler) setDirectory(ctx context.Context, arg Argument, value string) error {
	buf := bytes.NewBuffer(nil)
	dir := os.DirFS(value)

	if err := tarfs.Write(buf, dir); err != nil {
		return fmt.Errorf("error creating tar.gz for directory state: %w", err)
	}

	// Upload the buffer
	key := path.Join(s.BasePath, fmt.Sprintf("%s.tar.gz", stringutil.Slugify(arg.Key)))

	if err := s.Storage.PutObject(ctx, s.Bucket, key, buf); err != nil {
		return err
	}

	// Store the path to the tarball in the state
	return s.setValue(ctx, arg, key)
}

// setUnpackagedDirectory stores the path of a directory in the state without packaging it or uploading its contents.
// This should only be used for things that can be assumed are always available in the environment, like the source code, or elements in the source.
func (f *ObjectStorageHandler) setUnpackagedDirectory(ctx context.Context, arg Argument, value string) error {
	info, err := os.Stat(value)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("directory '%s' does not exist", value)
	}

	return f.setValue(ctx, arg, value)
}

func (s *ObjectStorageHandler) SetDirectory(ctx context.Context, arg Argument, path string) error {
	if arg.Type == ArgumentTypeFS {
		return s.setDirectory(ctx, arg, path)
	}
	return s.setUnpackagedDirectory(ctx, arg, path)
}
