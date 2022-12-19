package pipeline

import (
	"context"
	"fmt"

	"github.com/grafana/scribe/pipeline/dag"
)

// WalkFunc is implemented by the pipeline 'Clients'. This function is executed for each Step.
type StepWalkFunc func(context.Context, Step) error

// PipelineWalkFunc is implemented by the pipeline 'Clients'. This function is executed for each pipeline.
// This function follows the same rules for pipelines as the StepWalker func does for pipelines. If multiple pipelines are provided in the steps argument,
// then those pipelines are intended to be executed in parallel.
type PipelineWalkFunc func(context.Context, Pipeline) error

func (c *Collection) WalkPipelines(ctx context.Context, wf PipelineWalkFunc) error {
	if err := c.Graph.BreadthFirstSearch(0, c.pipelineVisitFunc(ctx, wf)); err != nil {
		return err
	}
	return nil
}

func (c *Collection) WalkSteps(ctx context.Context, pipelineID int64, wf StepWalkFunc) error {
	node, err := c.Graph.Node(pipelineID)
	if err != nil {
		return fmt.Errorf("could not find pipeline '%d'. %w", pipelineID, err)
	}

	pipeline := node.Value
	return pipeline.Graph.BreadthFirstSearch(0, func(n *dag.Node[Step]) error {
		if n.ID == 0 {
			return nil
		}

		return wf(ctx, n.Value)
	})
}
