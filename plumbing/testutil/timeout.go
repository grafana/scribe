package testutil

import (
	"fmt"
	"testing"
	"time"
)

// WithTimeout adds a timeout to the test function (f). If the `-timeout` flag is provided, then `f` will be called without a timeout and the `go test` command will handle the deadline.
func WithTimeout(d time.Duration, f func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		if _, ok := t.Deadline(); ok {
			f(t)
			return
		}

		go func() {
			<-time.After(d)
			panic(fmt.Sprintf("timeout '%s' exceeded", d))
		}()

		f(t)
	}
}
