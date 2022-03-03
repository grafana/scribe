package syncutil

import (
	"context"
	"fmt"
	"sync"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

// WaitGroup is a wrapper around a sync.WaitGroup that runs the actions of a list of steps, handles errors, and watches for context cancellation.
type WaitGroup struct {
	wg    sync.WaitGroup
	steps []pipeline.Step
}

// Add adds a new StepAction to the waitgroup. The provided function will be run in parallel with all other added functions.
func (wg *WaitGroup) Add(f pipeline.Step) {
	wg.steps = append(wg.steps, f)
}

// Wait runs all provided functions (via Add(...)) and runs them in parallel and waits for them to finish.
// If they are not all finished before the provided timeout (via NewWaitGroup), then an error is returned.
// If any functions return an error, the first error encountered is returned.
func (wg *WaitGroup) Wait(ctx context.Context, opts pipeline.ActionOpts) error {
	var (
		doneChan = make(chan bool)
		errChan  = make(chan error)
	)

	wg.wg.Add(len(wg.steps))

	for _, v := range wg.steps {
		go func(v pipeline.Step) {
			if err := v.Action(ctx, opts); err != nil {
				errChan <- err
			}

			wg.wg.Done()
		}(v)
	}

	go func() {
		wg.wg.Wait()
		doneChan <- true
	}()

	select {
	case <-ctx.Done():
		return context.Canceled
	case <-doneChan:
		return nil
	case err := <-errChan:
		return fmt.Errorf("error encountered in execution: %w", err)
	}
}

func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		wg:    sync.WaitGroup{},
		steps: []pipeline.Step{},
	}
}
