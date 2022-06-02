package yarn

import "github.com/grafana/scribe/plumbing/pipeline"

func Install() pipeline.Action {
	return NewStep("install")
}
