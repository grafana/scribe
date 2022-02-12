package yarn

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

func Install() pipeline.StepAction {
	return NewStep("install")
}
