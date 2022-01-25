package git

import (
	"net/url"

	"pkg.grafana.com/shipwright/v1/plumbing/config"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type CloneOpts struct {
	URL    *url.URL
	Folder string
	Ref    string
}

type Client struct {
	Configurer config.Configurer

	// Opts are provided to the Shipwright client (like the Drone client)
	// but most options could be valuable here, like "version"
	Opts *types.CommonOpts
}

func New(configurer config.Configurer, opts *types.CommonOpts) Client {
	return Client{
		Configurer: configurer,
		Opts:       opts,
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
