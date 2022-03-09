package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func NoOpAction(name string, duration time.Duration) pipeline.StepAction {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		f, err := os.Open(filepath.Join("demo", "complex", "logs", name+".log"))
		if err != nil {
			return err
		}

		time.Sleep(duration)

		io.ReadAll(io.TeeReader(f, opts.Stdout))

		return nil
	}
}
