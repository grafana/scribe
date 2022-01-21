package git

import (
	"net/url"

	"pkg.grafana.com/shipwright/v1/plumbing/config"
)

type CloneOpts struct {
	URL    *url.URL
	Folder string
	Ref    string
}

type Client struct {
	Configurer config.Configurer
}

func New(configurer config.Configurer) Client {
	return Client{
		Configurer: configurer,
	}
}

type DescribeOpts struct {
	Tags   bool
	Dirty  bool
	Always bool
}

func (c *Client) Describe(opts *DescribeOpts) string {
	return ""
}
