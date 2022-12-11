package state

import (
	"io"
	"io/fs"
	"os"
)

// NoOpHandler is a Handler that does nothing and is only used in tests where state is ignored but should never be nil.
type NoOpHandler struct{}

func NewNoOpHandler() *NoOpHandler { return &NoOpHandler{} }

// Reader functions
func (n *NoOpHandler) Exists(arg Argument) (bool, error)               { return false, nil }
func (n *NoOpHandler) GetString(arg Argument) (string, error)          { return "", nil }
func (n *NoOpHandler) GetInt64(arg Argument) (int64, error)            { return 0, nil }
func (n *NoOpHandler) GetFloat64(arg Argument) (float64, error)        { return 0.0, nil }
func (n *NoOpHandler) GetBool(arg Argument) (bool, error)              { return false, nil }
func (n *NoOpHandler) GetFile(arg Argument) (*os.File, error)          { return nil, nil }
func (n *NoOpHandler) GetDirectory(arg Argument) (fs.FS, error)        { return nil, nil }
func (n *NoOpHandler) GetDirectoryString(arg Argument) (string, error) { return "", nil }

// Writer functions
func (n *NoOpHandler) SetString(arg Argument, val string) error                { return nil }
func (n *NoOpHandler) SetInt64(arg Argument, val int64) error                  { return nil }
func (n *NoOpHandler) SetFloat64(arg Argument, val float64) error              { return nil }
func (n *NoOpHandler) SetBool(arg Argument, val bool) error                    { return nil }
func (n *NoOpHandler) SetFile(arg Argument, path string) error                 { return nil }
func (n *NoOpHandler) SetFileReader(arg Argument, r io.Reader) (string, error) { return "", nil }
func (n *NoOpHandler) SetDirectory(arg Argument, dir string) error             { return nil }
