package pipeline_test

import (
	"testing"

	"github.com/grafana/scribe/pipeline"
)

func TestStepIsBackground(t *testing.T) {
	step := pipeline.NamedStep("test step", pipeline.DefaultAction)
	step.Type = pipeline.StepTypeBackground

	if step.IsBackground() != true {
		t.Fatal("step.IsBackground should return true if the step.Type is pipeline.StepTypeBackground")
	}
}
