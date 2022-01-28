package make

import (
	"log"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

type Client struct{}

func (c *Client) Target(name string) types.StepAction {
	return func(types.ActionOpts) error {
		log.Println("make", name)
		return nil
	}
}
