package golang

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

func ModDownload() pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}
