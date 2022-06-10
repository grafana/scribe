package main

import "github.com/grafana/scribe/plumbing"

type ImageData struct {
	Version string
}

type Image struct {
	Name       string
	Dockerfile string
	Context    string
}

func (i Image) Tag(version string) string {
	// hack: if the image doesn't have a name then it must be the default one!
	name := plumbing.DefaultImage(version)

	if i.Name != "" {
		name = plumbing.SubImage(i.Name, version)
	}

	return name
}

// ScribeImage has to be built before its derivitive images.
var ScribeImage = Image{
	Name:       "",
	Dockerfile: "./ci/docker/scribe.Dockerfile",
	Context:    ".",
}

// Images is a list of images derived from the ScribeImage
var Images = []Image{
	{
		Name:       "git",
		Dockerfile: "./ci/docker/scribe.git.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "go",
		Dockerfile: "./ci/docker/scribe.go.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "node",
		Dockerfile: "./ci/docker/scribe.node.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "docker",
		Dockerfile: "./ci/docker/scribe.docker.Dockerfile",
		Context:    ".",
	},
}
