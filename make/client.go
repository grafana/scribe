package make

import "pkg.grafana.com/shipwright/v1/types"

type Client struct{}

func (c *Client) Target(name string) types.Step {
	return func() error {
		return nil
	}
}
