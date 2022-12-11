package state

import (
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

func (o *Observer) CondFor(arg Argument) *sync.Cond {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if val, ok := o.conds[arg]; ok {
		return val
	}

	o.conds[arg] = sync.NewCond(&sync.Mutex{})
	return o.conds[arg]
}

func (o *Observer) Notify(arg Argument) {
	o.mtx.Lock()
	defer o.mtx.Unlock()
	if v, ok := o.conds[arg]; ok {
		v.L.Lock()
		v.Broadcast()
		v.L.Unlock()
	}
}

// Reader functions
func (o *Observer) Exists(arg Argument) (bool, error)        { return o.h.Exists(arg) }
func (o *Observer) GetString(arg Argument) (string, error)   { return o.h.GetString(arg) }
func (o *Observer) GetInt64(arg Argument) (int64, error)     { return o.h.GetInt64(arg) }
func (o *Observer) GetFloat64(arg Argument) (float64, error) { return o.h.GetFloat64(arg) }
func (o *Observer) GetBool(arg Argument) (bool, error)       { return o.h.GetBool(arg) }
func (o *Observer) GetFile(arg Argument) (*os.File, error)   { return o.h.GetFile(arg) }
func (o *Observer) GetDirectory(arg Argument) (fs.FS, error) { return o.h.GetDirectory(arg) }
func (o *Observer) GetDirectoryString(arg Argument) (string, error) {
	return o.h.GetDirectoryString(arg)
}

// Writer functions
func (o *Observer) SetString(arg Argument, val string) error {
	defer o.Notify(arg)
	return o.h.SetString(arg, val)
}
func (o *Observer) SetInt64(arg Argument, val int64) error {
	defer o.Notify(arg)
	return o.h.SetInt64(arg, val)
}
func (o *Observer) SetFloat64(arg Argument, val float64) error {
	defer o.Notify(arg)
	return o.h.SetFloat64(arg, val)
}
func (o *Observer) SetBool(arg Argument, val bool) error {
	defer o.Notify(arg)
	return o.h.SetBool(arg, val)
}
func (o *Observer) SetFile(arg Argument, path string) error {
	defer o.Notify(arg)
	return o.h.SetFile(arg, path)
}
func (o *Observer) SetFileReader(arg Argument, r io.Reader) (string, error) {
	defer o.Notify(arg)
	return o.h.SetFileReader(arg, r)
}
func (o *Observer) SetDirectory(arg Argument, dir string) error {
	defer o.Notify(arg)
	return o.h.SetDirectory(arg, dir)
}
