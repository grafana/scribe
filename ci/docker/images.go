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

func str(val string) *string {
	return &val
}

type ImageData struct {
	Version string
}

func (i Image) Tag() (string, error) {
	v, err := version()
	if err != nil {
		return "", err
	}

	// hack: if the image doesn't have a name then it must be the default one!
	name := plumbing.DefaultImage(v)

	if i.Name != "" {
		name = plumbing.SubImage(i.Name, v)
	}

	return name, err
}

func version() (string, error) {
	// git config --global --add safe.directory * is needed to resolve the restriction introduced by CVE-2022-24765.
	out, err := exec.Command("git", "config", "--global", "--add", "safe.directory", "*").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running command 'git config --global --add safe.directory *' resulted in error '%w'. Output: '%s'", err, string(out))
	}

	version, err := exec.Command("git", "describe", "--tags", "--dirty", "--always").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running command 'git describe --tags --dirty --always' resulted in the error '%w'. Output: '%s'", err, string(version))
	}

	return strings.TrimSpace(string(version)), nil
}

type Image struct {
	Name       string
	Dockerfile string
	Context    string
}

func (i Image) BuildStep(sw *shipwright.Shipwright[pipeline.Action]) pipeline.Step[pipeline.Action] {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		tag, err := i.Tag()
		if err != nil {
			return err
		}

		v, err := version()
		if err != nil {
			return err
		}

		opts.Logger.Infoln("Building", i.Dockerfile, "with tag", tag)
		return docker.Build(ctx, docker.BuildOptions{
			Names:      []string{tag},
			Dockerfile: i.Dockerfile,
			ContextDir: i.Context,
			Args: map[string]*string{
				"VERSION": str(v),
			},
			Stdout: opts.Stdout,
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS).
		WithImage(plumbing.SubImage("docker", sw.Version))
}

func (i Image) PushStep(sw *shipwright.Shipwright[pipeline.Action]) pipeline.Step[pipeline.Action] {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {
		tag, err := i.Tag()
		if err != nil {
			return err
		}

		auth, err := opts.State.Get(ArgumentDockerAuthToken.Key)
		if err != nil {
			return err
		}

		opts.Logger.Infoln("Pushing", tag)
		return docker.Push(ctx, docker.PushOpts{
			Name:      tag,
			Registry:  plumbing.DefaultRegistry(),
			AuthToken: auth,
			Stdout:    opts.Stdout,
		})
	}

	return pipeline.NewStep(action).
		WithArguments(pipeline.ArgumentSourceFS, pipeline.ArgumentDockerSocketFS, ArgumentDockerAuthToken).
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

func BuildSteps(sw *shipwright.Shipwright[pipeline.Action], images []Image) []pipeline.Step[pipeline.Action] {
	steps := make([]pipeline.Step[pipeline.Action], len(images))

	for i, image := range images {
		steps[i] = image.BuildStep(sw).WithName(fmt.Sprintf("build %s image", image.Name))
	}

	return steps
}

func PushSteps(sw *shipwright.Shipwright[pipeline.Action], images []Image) []pipeline.Step[pipeline.Action] {
	steps := make([]pipeline.Step[pipeline.Action], len(images))

	for i, image := range images {
		steps[i] = image.PushStep(sw).WithName(fmt.Sprintf("push %s", image.Name))
	}

	return steps
}
