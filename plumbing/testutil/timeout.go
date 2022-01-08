package testutil

import (
	"testing"
	"time"
)

func WithTimeout(d time.Duration, f func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		timer := time.NewTimer(d)
		doneChan := make(chan bool)
		go func() {
			f(t)
			doneChan <- true
		}()

		select {
		case <-doneChan:
			timer.Stop()
			return
		case <-timer.C:
			t.Fatalf("'%s' exceeded", d)
		}
	}
}
