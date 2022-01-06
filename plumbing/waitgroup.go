package plumbing

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type wgfunc func() error

// WaitGroup is a wrapper around a sync.WaitGroup that handles errors
type WaitGroup struct {
	timeout time.Duration
	wg      sync.WaitGroup
	funcs   []wgfunc
}

func (wg *WaitGroup) Add(f func() error) {
	wg.funcs = append(wg.funcs, f)
}

func (wg *WaitGroup) Wait() error {
	var (
		doneChan = make(chan bool)
		errChan  = make(chan error)
	)

	t := time.NewTimer(wg.timeout)

	log.Println("got", len(wg.funcs))
	wg.wg.Add(len(wg.funcs))

	for _, v := range wg.funcs {
		go func(v wgfunc) {
			if err := v(); err != nil {
				errChan <- err
			}

			wg.wg.Done()
		}(v)
	}

	go func() {
		wg.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		log.Println("Done running step(s) without errors")
	case err := <-errChan:
		return fmt.Errorf("error encountered in execution: %w", err)
	case <-t.C:
		return errors.New("time out")
	}
	return nil
}

func NewWaitGroup(timeout time.Duration) *WaitGroup {
	return &WaitGroup{
		timeout: timeout,
		wg:      sync.WaitGroup{},
		funcs:   []wgfunc{},
	}
}
