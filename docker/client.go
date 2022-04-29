package docker

import (
	"sync"

	"github.com/docker/docker/client"
)

var cli *client.Client

func initClient() {
	cli = newClient()
}

func dockerClient() *client.Client {
	once := &sync.Once{}
	once.Do(initClient)
	return cli
}

func newClient() *client.Client {
	c, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return c
}
