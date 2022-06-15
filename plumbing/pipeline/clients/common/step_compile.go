package common

import (
	"context"

	"github.com/grafana/scribe/plumbing/pipeline"
)

func compilePipeline(ctx context.Context, opts pipeline.ActionOpts) error {
	return nil
}

func StepCompilePipeline() pipeline.Step {
	return pipeline.NewStep(compilePipeline).
		WithImage("golang:1.18").
		WithName("builtin-compile-pipeline").
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentPipelineFS)
}
