package yarn

import "github.com/grafana/shipwright/plumbing/pipeline"

func Install() pipeline.StepAction {
	return NewStep("install")
}
