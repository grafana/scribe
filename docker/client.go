package docker

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

type Client struct {
	CommonOpts *pipeline.CommonOpts
}

func New(c *pipeline.CommonOpts) Client {
	return Client{
		CommonOpts: c,
	}
}
