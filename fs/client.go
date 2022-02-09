package fs

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

type Client struct{}

func (c *Client) Replace(file string, content string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}

func (c *Client) ReplaceString(file string, content string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		return nil
	}
}
