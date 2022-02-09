package syncutil

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// WaitGroup is a wrapper around a sync.WaitGroup that handles errors
type WaitGroup struct {
	timeout time.Duration
	wg      sync.WaitGroup
	steps   []types.Step
}

func (w *WaitGroup) StepNames() []string {
	if w.steps == nil {
		return []string{}
	}

	names := make([]string, len(w.steps))

	for i, v := range w.steps {
		names[i] = v.Name
	}

	return names
}

// Add adds a new StepAction to the waitgroup. The provided function will be run in parallel with all other added functions.
func (wg *WaitGroup) Add(f types.Step) {
	wg.steps = append(wg.steps, f)
}

// Wait runs all provided functions (via Add(...)) and runs them in parallel and waits for them to finish.
// If they are not all finished before the provided timeout (via NewWaitGroup), then an error is returned.
// If any functions return an error, the first error encountered is returned.
func (wg *WaitGroup) Wait(ctx context.Context, opts types.ActionOpts) error {
	var (
		doneChan = make(chan bool)
		errChan  = make(chan error)
	)

	t := time.NewTimer(wg.timeout)

	wg.wg.Add(len(wg.steps))

	for _, v := range wg.steps {
		go func(v types.Step) {
			if err := v.Action(opts); err != nil {
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
		return nil
	case err := <-errChan:
		return fmt.Errorf("error encountered in execution: %w", err)
	case <-t.C:
		return errors.New("time out")
	}
}

func NewWaitGroup(timeout time.Duration) *WaitGroup {
	return &WaitGroup{
		timeout: timeout,
		wg:      sync.WaitGroup{},
		steps:   []types.Step{},
	}
}
