package main

import (
	docker "github.com/fsouza/go-dockerclient"
)

func Client() *docker.Client {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	return client
}
