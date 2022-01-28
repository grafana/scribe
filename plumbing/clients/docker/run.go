package docker

import (
	"io"
	"os/exec"

	"pkg.grafana.com/shipwright/v1/plumbing/plog"
)

type RunOpts struct {
	Image   string
	Command string
	Volumes []string
	Args    []string

	Stdout io.Writer
	Stderr io.Writer
}

func Run(opts RunOpts) error {
	volumes := []string{}
	for _, v := range opts.Volumes {
		volumes = append(volumes, "-v", v)
	}

	args := []string{"run", "--rm"}
	args = append(args, volumes...)
	args = append(args, opts.Image)
	args = append(args, opts.Command)
	args = append(args, opts.Args...)

	plog.Infoln("Running command", "docker", args)
	cmd := exec.Command("docker", args...)
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	return cmd.Run()
}
