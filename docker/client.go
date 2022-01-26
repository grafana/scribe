package docker

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Client struct {
	CommonOpts *types.CommonOpts
}

func New(c *types.CommonOpts) Client {
	return Client{
		CommonOpts: c,
	}
}
