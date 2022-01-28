package yarn

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Client struct {
}

func (c *Client) Install() types.StepAction {
	return NewStep("install")
}
