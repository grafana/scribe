package main

import (
	"context"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/exec"
	"github.com/grafana/scribe/pipeline"
)

func echo(ctx context.Context, opts pipeline.ActionOpts) error {
	return exec.RunCommandWithOpts(ctx, exec.RunOpts{
		Name:   "/bin/sh",
		Args:   []string{"-c", `sleep 10; echo "hello ?"`},
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
	})
}

func StepEcho() pipeline.Step {
	return pipeline.NewStep(echo).WithImage("ubuntu:latest")
}

func main() {
	sw := scribe.New("test-pipeline")
	defer sw.Done()

	sw.Add(StepEcho())
}
