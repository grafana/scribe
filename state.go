package scribe

import (
	"context"

	"dagger.io/dagger"
)

type StateReader interface {
	Exists(context.Context, Argument) (bool, error)
	GetString(context.Context, Argument) (string, error)
	GetInt64(context.Context, Argument) (int64, error)
	GetDirectory(context.Context, Argument) (*dagger.Directory, error)
	GetFile(context.Context, Argument) (*dagger.File, error)
}

type StateWriter interface {
	SetString(Argument, string) error
	SetInt64(Argument, int64) error
	SetDirectory(Argument, *dagger.Directory) error
	SetFile(Argument, *dagger.File) error
}

type StateHandler interface {
	StateReader
	StateWriter
}

type State struct {
	Handler StateHandler
}

func (s *State) Exists(ctx context.Context, arg Argument) (bool, error) {
	return s.Handler.Exists(ctx, arg)
}

func (s *State) GetString(ctx context.Context, arg Argument) (string, error) {
	return s.Handler.GetString(ctx, arg)
}

func (s *State) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	return s.Handler.GetInt64(ctx, arg)
}

func (s *State) GetDirectory(ctx context.Context, arg Argument) (*dagger.Directory, error) {
	return s.Handler.GetDirectory(ctx, arg)
}

func (s *State) GetFile(ctx context.Context, arg Argument) (*dagger.File, error) {
	return s.Handler.GetFile(ctx, arg)
}

func (s *State) SetString(arg Argument, val string) error {
	return s.Handler.SetString(arg, val)
}

func (s *State) SetInt64(arg Argument, val int64) error {
	return s.Handler.SetInt64(arg, val)
}

func (s *State) SetDirectory(arg Argument, val *dagger.Directory) error {
	return s.Handler.SetDirectory(arg, val)
}

func (s *State) SetFile(arg Argument, val *dagger.File) error {
	return s.Handler.SetFile(arg, val)
}
