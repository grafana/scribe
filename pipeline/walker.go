package pipeline

import "context"

// WalkFunc is implemented by the pipeline 'Clients'. This function is executed for each Step.
type StepWalkFunc func(context.Context, Step) error

// PipelineWalkFunc is implemented by the pipeline 'Clients'. This function is executed for each pipeline.
// This function follows the same rules for pipelines as the StepWalker func does for pipelines. If multiple pipelines are provided in the steps argument,
// then those pipelines are intended to be executed in parallel.
type PipelineWalkFunc func(context.Context, ...Pipeline) error

// Walker is an interface that collections of steps should satisfy.
type Walker interface {
	WalkSteps(context.Context, int64, StepWalkFunc) error
	WalkPipelines(context.Context, PipelineWalkFunc) error
}
