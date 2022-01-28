package syncutil

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// WaitGroup is a wrapper around a sync.WaitGroup that handles errors
type WaitGroup struct {
	timeout time.Duration
	wg      sync.WaitGroup
	funcs   []types.StepAction
}

// Add adds a new StepAction to the waitgroup. The provided function will be run in parallel with all other added functions.
func (wg *WaitGroup) Add(f types.StepAction) {
	wg.funcs = append(wg.funcs, f)
}

// Wait runs all provided functions (via Add(...)) and runs them in parallel and waits for them to finish.
// If they are not all finished before the provided timeout (via NewWaitGroup), then an error is returned.
// If any functions return an error, the first error encountered is returned.
func (wg *WaitGroup) Wait(opts types.ActionOpts) error {
	var (
		doneChan = make(chan bool)
		errChan  = make(chan error)
	)

	t := time.NewTimer(wg.timeout)

	wg.wg.Add(len(wg.funcs))

	for _, v := range wg.funcs {
		go func(v types.StepAction) {
			if err := v(opts); err != nil {
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
		funcs:   []types.StepAction{},
	}
}
