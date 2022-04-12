package main

import (
	"context"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/opentracing/opentracing-go"
)

func NoOpAction(name string, duration time.Duration) pipeline.Action {
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

func IntegrationTest(variant string, duration time.Duration) pipeline.Action {
	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		d := int64(duration.Seconds()) / 2
		tests := []string{"fs", "docker", "exec", "git", "golang", "makefile", "yarn"}
		parent := opentracing.SpanFromContext(ctx)

		for _, test := range tests {
			span, _ := opentracing.StartSpanFromContextWithTracer(ctx, opts.Tracer, test, opentracing.ChildOf(parent.Context()))
			span.SetTag("job", "shipwright")
			l := log.New(opts.Stdout, variant, 0)
			l.Printf("Testing '%s' package with '%s'...", test, variant)

			r := rand.Int63n(d)
			d -= r
			time.Sleep(time.Duration(r) * time.Second)
			span.Finish()
		}

		return nil
	}
}
