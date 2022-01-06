package make

import (
	"log"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type Client struct{}

func (c *Client) Target(name string) types.Step {
	return func() error {
		log.Println("make", name)
		return nil
	}
}
