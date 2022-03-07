package docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/docker"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
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

	return strings.TrimSpace(string(version)), nil
}

func (i Image) BuildStep(sw shipwright.Shipwright) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		v, err := version()
		if err != nil {
			return err
		}

		// hack: if the image doesn't have a name then it must be the default one!
		name := plumbing.DefaultImage(v)

		if i.Name != "" {
			name = plumbing.SubImage(i.Name, v)
		}

		return docker.BuildWithArgs(name, i.Dockerfile, i.Context, fmt.Sprintf("VERSION=%s", v)).Action(ctx, opts)
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
