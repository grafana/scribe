package git

import (
	"log"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type DescribeOpts struct {
	Tags   bool
	Dirty  bool
	Always bool
}

type Client struct{}

func (c *Client) Describe(opts *DescribeOpts) string {
	return ""
}

func (c *Client) Clone() types.Step {
	return func() error {
		log.Println("git clone ...")
		return nil
	}
}
