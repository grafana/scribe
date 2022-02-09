package docker

import (
	"pkg.grafana.com/shipwright/v1/exec"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
)

func (c Client) BuildWithArgs(name, path, context string, arg ...string) pipeline.Step {
	args := []string{
		"build",
		"-f", path,
		"-t", name,
	}

	for _, v := range arg {
		args = append(args, "--build-arg", v)
	}

	args = append(args, context)
	return pipeline.NewStep(exec.Run("docker", args...))
}

func (c Client) Build(name, path, context string) pipeline.Step {
	return c.BuildWithArgs(name, path, context).WithArguments(pipeline.ArgumentDockerSocketFS)
}
