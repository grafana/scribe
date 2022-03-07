package docker

import "github.com/grafana/shipwright/plumbing/pipeline"

type Client struct {
	CommonOpts *pipeline.CommonOpts
}

func New(c *pipeline.CommonOpts) Client {
	return Client{
		CommonOpts: c,
	}
}
