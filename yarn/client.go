package yarn

import "github.com/grafana/shipwright/plumbing/pipeline"

func Install() pipeline.Action {
	return NewStep("install")
}
