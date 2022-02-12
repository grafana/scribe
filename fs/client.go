package fs

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

func Replace(file string, content string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}

func ReplaceString(file string, content string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}
