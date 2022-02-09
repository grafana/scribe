package docker

import (
	"fmt"
	"os/exec"
	"strings"

	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

type ImageData struct {
	Version string
}

type Image struct {
	Name       string
	Dockerfile string
	Context    string
}

func version() (string, error) {
	version, err := exec.Command("git", "describe", "--tags", "--dirty", "--always").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running command 'git describe --tags --dirty --always' resulted in the error '%w'. Output: '%s'", err, string(version))
	}

	return string(version), nil
}

func (i Image) BuildStep(sw shipwright.Shipwright) pipeline.Step {
	action := func(opts pipeline.ActionOpts) error {
		v, err := version()
		if err != nil {
			return err
		}

		v = strings.TrimSpace(v)

		// hack: if the image doesn't have a name then it must be the default one!
		name := plumbing.DefaultImage(v)

		if i.Name != "" {
			name = plumbing.SubImage(i.Name, v)
		}

		return sw.Docker.BuildWithArgs(name, i.Dockerfile, i.Context, fmt.Sprintf("VERSION=%s", v)).Action(opts)
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS).
		WithImage(plumbing.SubImage("docker", sw.Version))
}

// ShipwrightImage has to be built before its derivitive images.
var ShipwrightImage = Image{
	Dockerfile: "./ci/docker/shipwright.Dockerfile",
	Context:    ".",
}

// Images is a list of images derived from the ShipwrightImage
var Images = []Image{
	{
		Name:       "git",
		Dockerfile: "./ci/docker/shipwright.git.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "go",
		Dockerfile: "./ci/docker/shipwright.go.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "node",
		Dockerfile: "./ci/docker/shipwright.node.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "docker",
		Dockerfile: "./ci/docker/shipwright.docker.Dockerfile",
		Context:    ".",
	},
}

func Steps(sw shipwright.Shipwright, images []Image) []pipeline.Step {
	steps := make([]pipeline.Step, len(images))

	for i, image := range images {
		steps[i] = image.BuildStep(sw).WithName(fmt.Sprintf("build %s image", image.Name))
	}

	return steps
}
