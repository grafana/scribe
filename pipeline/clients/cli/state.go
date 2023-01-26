package cli

import (
	"context"
	"io"
	"os"

	"github.com/grafana/scribe/state"
)

type StateHandler struct{}

func (n *StateHandler) SetString(ctx context.Context, arg state.Argument, val string) error {
	return nil
}
func (n *StateHandler) SetInt64(ctx context.Context, arg state.Argument, val int64) error { return nil }
func (n *StateHandler) SetFloat64(ctx context.Context, arg state.Argument, val float64) error {
	return nil
}
func (n *StateHandler) SetBool(ctx context.Context, arg state.Argument, val bool) error { return nil }
func (n *StateHandler) SetFile(ctx context.Context, arg state.Argument, path string) error {
	return nil
}
func (n *StateHandler) SetFileReader(ctx context.Context, arg state.Argument, r io.Reader) (string, error) {
	file, err := os.CreateTemp("", "*")
	if err != nil {
		return "", err
	}

	defer file.Close()
	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (n *StateHandler) SetDirectory(ctx context.Context, arg state.Argument, dir string) error {
	return nil
}
