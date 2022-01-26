package docker

import (
	"fmt"
	"os/exec"
	"strings"

	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
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

func (i Image) BuildStep(sw shipwright.Shipwright) types.Step {
	action := func() error {
		v, err := version()
		if err != nil {
			return err
		}

		v = strings.TrimSpace(v)

		return sw.Docker.BuildWithArgs(fmt.Sprintf("%s:%s", i.Name, v), i.Dockerfile, i.Context, fmt.Sprintf("VERSION=%s", v)).Action()
	}

	return types.NewStep(action)
}

// ShipwrightImage has to be built before its derivitive images.
var ShipwrightImage = Image{
	Name:       "shipwright",
	Dockerfile: "./ci/docker/shipwright.Dockerfile",
	Context:    ".",
}

// Images is a list of images derived from the ShipwrightImage
var Images = []Image{
	{
		Name:       "shipwright/git",
		Dockerfile: "./ci/docker/shipwright.git.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "shipwright/go",
		Dockerfile: "./ci/docker/shipwright.go.Dockerfile",
		Context:    ".",
	},
	{
		Name:       "shipwright/node",
		Dockerfile: "./ci/docker/shipwright.node.Dockerfile",
		Context:    ".",
	},
}

func Steps(sw shipwright.Shipwright, images []Image) []types.Step {
	steps := make([]types.Step, len(images))

	for i, image := range images {
		steps[i] = image.BuildStep(sw).WithName(fmt.Sprintf("build %s image", image.Name))
	}

	return steps
}
