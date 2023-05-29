package scribe

import (
	"sync"

	"dagger.io/dagger"
)

type ArgumentType int

const (
	ArgumentTypeInt64 ArgumentType = iota
	ArgumentTypeFloat64
	ArgumentTypeString
	ArgumentTypeBool
	ArgumentTypeFile
	ArgumentTypeDirectory
)

type Argument struct {
	Key   string
	Type  ArgumentType
	value interface{}
	c     *sync.Cond
	mtx   *sync.Mutex
}

func (a Argument) C() *sync.Cond {
	return a.c
}

func (a Argument) Int64() (int64, error) {
	return 0, nil
}

func (a Argument) Float64() (float64, error) {
	return 0.0, nil
}

func (a Argument) String() (string, error) {
	return "", nil
}

func (a Argument) Bool() (bool, error) {
	return false, nil
}

func (a Argument) File() (*dagger.File, error) {
	return nil, nil
}

func (a Argument) Directory() (*dagger.Directory, error) {
	return nil, nil
}

func NewStringArgument(key string) Argument {
	return Argument{
		Key:  key,
		Type: ArgumentTypeString,
	}
}
