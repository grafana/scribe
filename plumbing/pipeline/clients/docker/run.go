package docker

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

// PipelineVolumePath refers to the path where the compiled pipeline is mounted in the container
const PipelineVolumePath = "/var/pipeline"

type RunOpts struct {
	PipelinePath string
	Image        string
	Command      string
	Volumes      []string
	Args         []string

	Stdout io.Writer
	Stderr io.Writer
}

func RunArgs(opts RunOpts) []string {
	volumes := []string{
		"-v", fmt.Sprintf("%s:%s", opts.PipelinePath, "/var/pipeline"),
	}

	for _, v := range opts.Volumes {
		volumes = append(volumes, "-v", v)
	}

	args := []string{"run", "--rm"}
	args = append(args, volumes...)
	args = append(args, opts.Image)
	args = append(args, opts.Command)
	args = append(args, opts.Args...)

	return args
}

func Run(ctx context.Context, opts RunOpts) error {
	cmd := exec.CommandContext(ctx, "docker", RunArgs(opts)...)
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	return cmd.Run()
}
