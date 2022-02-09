package makefile

import (
	"log"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

type Client struct{}

func (c *Client) Target(name string) pipeline.StepAction {
	return func(pipeline.ActionOpts) error {
		log.Println("make", name)
		return nil
	}
}
