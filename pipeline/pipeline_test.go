package pipeline_test

import (
	"testing"

	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

func TestBuildEdges(t *testing.T) {
	t.Run("An edge from the root node should be created if the pipeline and step depend on the same argument", func(t *testing.T) {
		arg := state.NewStringArgument("shared-arg")
		p := pipeline.New("test-pipeline", 1).Requires(arg)
		step1 := pipeline.NoOpStep.WithName("step 1").Requires(arg)
		step1.ID = 5
		if err := p.AddSteps(step1); err != nil {
			t.Fatal(err)
		}

		rootArgs := []state.Argument{}
		if err := p.BuildEdges(rootArgs...); err != nil {
			t.Fatal(err)
		}
	})
}
