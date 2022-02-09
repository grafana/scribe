package yarn

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

type Client struct {
}

func (c *Client) Install() pipeline.StepAction {
	return NewStep("install")
}
