package state

import (
	"context"
	"io"
	"io/fs"
	"os"
)

// NoOpHandler is a Handler that does nothing and is only used in tests where state is ignored but should never be nil.
type NoOpHandler struct{}

func NewNoOpHandler() *NoOpHandler { return &NoOpHandler{} }

// Reader functions
func (n *NoOpHandler) Exists(ctx context.Context, arg Argument) (bool, error)        { return false, nil }
func (n *NoOpHandler) GetString(ctx context.Context, arg Argument) (string, error)   { return "", nil }
func (n *NoOpHandler) GetInt64(ctx context.Context, arg Argument) (int64, error)     { return 0, nil }
func (n *NoOpHandler) GetFloat64(ctx context.Context, arg Argument) (float64, error) { return 0.0, nil }
func (n *NoOpHandler) GetBool(ctx context.Context, arg Argument) (bool, error)       { return false, nil }
func (n *NoOpHandler) GetFile(ctx context.Context, arg Argument) (*os.File, error)   { return nil, nil }
func (n *NoOpHandler) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) { return nil, nil }
func (n *NoOpHandler) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	return "", nil
}

// Writer functions
func (n *NoOpHandler) SetString(ctx context.Context, arg Argument, val string) error   { return nil }
func (n *NoOpHandler) SetInt64(ctx context.Context, arg Argument, val int64) error     { return nil }
func (n *NoOpHandler) SetFloat64(ctx context.Context, arg Argument, val float64) error { return nil }
func (n *NoOpHandler) SetBool(ctx context.Context, arg Argument, val bool) error       { return nil }
func (n *NoOpHandler) SetFile(ctx context.Context, arg Argument, path string) error    { return nil }
func (n *NoOpHandler) SetFileReader(ctx context.Context, arg Argument, r io.Reader) (string, error) {
	return "", nil
}
func (n *NoOpHandler) SetDirectory(ctx context.Context, arg Argument, dir string) error { return nil }
