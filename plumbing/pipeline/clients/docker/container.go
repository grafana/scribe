package docker

import (
	"context"
	"io"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/grafana/scribe/plumbing/cmdutil"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/scribe/plumbing/stringutil"
)

type CreateStepContainerOpts struct {
	Configurer pipeline.Configurer
	Step       pipeline.Step
	Env        []string
	Network    *docker.Network
	Volumes    []*docker.Volume
	Mounts     []docker.HostMount
	Binary     string
	Pipeline   string
	BuildID    string
	Out        io.Writer
}

func CreateStepContainer(ctx context.Context, client *docker.Client, opts CreateStepContainerOpts) (*docker.Container, error) {
	cmd, err := cmdutil.StepCommand(cmdutil.CommandOpts{
		CompiledPipeline: opts.Binary,
		Path:             opts.Pipeline,
		Step:             opts.Step,
		BuildID:          opts.BuildID,
		State:            "file:///var/scribe-state/state.json",
	})

	if err != nil {
		return nil, err
	}

	createOpts, err := applyArguments(opts.Configurer, docker.CreateContainerOptions{
		Context: ctx,
		Name:    strings.Join([]string{"scribe", stringutil.Slugify(opts.Step.Name), stringutil.Random(8)}, "-"),
		Config: &docker.Config{
			Image:        opts.Step.Image,
			Cmd:          cmd,
			AttachStdout: true,
			AttachStderr: true,
			Env:          append(opts.Env, "GIT_CEILING_DIRECTORIES=/var/scribe"),
		},
		HostConfig: &docker.HostConfig{
			NetworkMode: opts.Network.Name,
			Mounts:      opts.Mounts,
		},
	}, opts.Step.Arguments)

	return client.CreateContainer(createOpts)
}

type RunOpts struct {
	Container  *docker.Container
	HostConfig *docker.HostConfig
	Stdout     io.Writer
	Stderr     io.Writer
}

func RunContainer(ctx context.Context, client *docker.Client, opts RunOpts) error {
	if err := client.StartContainerWithContext(opts.Container.ID, opts.HostConfig, ctx); err != nil {
		return err
	}

	if err := client.AttachToContainer(docker.AttachToContainerOptions{
		Container:    opts.Container.ID,
		OutputStream: opts.Stdout,
		ErrorStream:  opts.Stderr,
		Stream:       true,
		Stdout:       true,
		Stderr:       true,
	}); err != nil {
		return err
	}

	return nil
}
