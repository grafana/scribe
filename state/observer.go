package state

import (
	"context"
	"io"
	"io/fs"
	"os"
	"sync"
)

type Observer struct {
	h     Handler
	conds map[Argument]*sync.Cond
	mtx   *sync.Mutex
}

func NewObserver(h Handler) *Observer {
	return &Observer{
		h:     h,
		conds: make(map[Argument]*sync.Cond),
		mtx:   &sync.Mutex{},
	}
}

func (o *Observer) CondFor(ctx context.Context, arg Argument) *sync.Cond {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if val, ok := o.conds[arg]; ok {
		return val
	}

	o.conds[arg] = sync.NewCond(&sync.Mutex{})
	return o.conds[arg]
}

func (o *Observer) Notify(ctx context.Context, arg Argument) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if v, ok := o.conds[arg]; ok {
		v.L.Lock()
		v.Broadcast()
		v.L.Unlock()
	}
}

// Reader functions
func (o *Observer) Exists(ctx context.Context, arg Argument) (bool, error) {
	return o.h.Exists(ctx, arg)
}
func (o *Observer) GetString(ctx context.Context, arg Argument) (string, error) {
	return o.h.GetString(ctx, arg)
}
func (o *Observer) GetInt64(ctx context.Context, arg Argument) (int64, error) {
	return o.h.GetInt64(ctx, arg)
}
func (o *Observer) GetFloat64(ctx context.Context, arg Argument) (float64, error) {
	return o.h.GetFloat64(ctx, arg)
}
func (o *Observer) GetBool(ctx context.Context, arg Argument) (bool, error) {
	return o.h.GetBool(ctx, arg)
}
func (o *Observer) GetFile(ctx context.Context, arg Argument) (*os.File, error) {
	return o.h.GetFile(ctx, arg)
}
func (o *Observer) GetDirectory(ctx context.Context, arg Argument) (fs.FS, error) {
	return o.h.GetDirectory(ctx, arg)
}
func (o *Observer) GetDirectoryString(ctx context.Context, arg Argument) (string, error) {
	return o.h.GetDirectoryString(ctx, arg)
}

// Writer functions
func (o *Observer) SetString(ctx context.Context, arg Argument, val string) error {
	defer o.Notify(ctx, arg)
	return o.h.SetString(ctx, arg, val)
}
func (o *Observer) SetInt64(ctx context.Context, arg Argument, val int64) error {
	defer o.Notify(ctx, arg)
	return o.h.SetInt64(ctx, arg, val)
}
func (o *Observer) SetFloat64(ctx context.Context, arg Argument, val float64) error {
	defer o.Notify(ctx, arg)
	return o.h.SetFloat64(ctx, arg, val)
}
func (o *Observer) SetBool(ctx context.Context, arg Argument, val bool) error {
	defer o.Notify(ctx, arg)
	return o.h.SetBool(ctx, arg, val)
}
func (o *Observer) SetFile(ctx context.Context, arg Argument, path string) error {
	defer o.Notify(ctx, arg)
	return o.h.SetFile(ctx, arg, path)
}
func (o *Observer) SetFileReader(ctx context.Context, arg Argument, r io.Reader) (string, error) {
	defer o.Notify(ctx, arg)
	return o.h.SetFileReader(ctx, arg, r)
}
func (o *Observer) SetDirectory(ctx context.Context, arg Argument, dir string) error {
	defer o.Notify(ctx, arg)
	return o.h.SetDirectory(ctx, arg, dir)
}
