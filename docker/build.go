package docker

import (
	"github.com/grafana/shipwright/exec"

	"github.com/grafana/shipwright/plumbing/pipeline"
)

func BuildWithArgs(name, path, context string, arg ...string) pipeline.Step {
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

func Build(name, path, context string) pipeline.Step {
	return BuildWithArgs(name, path, context).WithArguments(pipeline.ArgumentDockerSocketFS)
}
