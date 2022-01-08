package fs

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Client struct{}

func (c *Client) Replace(file string, content string) types.StepAction {
	return func() error {
		return nil
	}
}

func (c *Client) ReplaceString(file string, content string) types.StepAction {
	return func() error {
		return nil
	}
}
