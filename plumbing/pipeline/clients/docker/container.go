package docker

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/grafana/scribe/docker"
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
	Mounts     []mount.Mount
	Binary     string
	Pipeline   string
	BuildID    string
	Out        io.Writer
}

func CreateStepContainer(ctx context.Context, cli client.APIClient, opts CreateStepContainerOpts) (*docker.Container, error) {
	cmd, err := cmdutil.StepCommand(cmdutil.CommandOpts{
		CompiledPipeline: opts.Binary,
		Path:             opts.Pipeline,
		Step:             opts.Step,
		BuildID:          opts.BuildID,
		State:            "/var/scribe-state/state.json",
	})

	if err != nil {
		return nil, err
	}

	createOpts, err := applyArguments(opts.Configurer, docker.CreateContainerOpts{
		Name:    strings.Join([]string{"scribe", stringutil.Slugify(opts.Step.Name), stringutil.Random(8)}, "-"),
		Image:   opts.Step.Image,
		Network: opts.Network,
		Mounts:  opts.Mounts,
		Command: cmd,
		Out:     opts.Out,
		Env:     append(opts.Env, "GIT_CEILING_DIRECTORIES=/var/scribe"),
	}, opts.Step.Arguments)

	return docker.CreateContainer(ctx, cli, createOpts)
}

// Value retrieves the configuration item the same way the CLI does; by looking in the argmap or by asking via stdin.
func (c *Client) Value(arg pipeline.Argument) (string, error) {
	switch arg.Type {
	//case pipeline.ArgumentTypeString:
	//	return cli.GetArgValue(c.Opts.Args, arg)
	case pipeline.ArgumentTypeFS:
		return GetVolumeValue(c.Opts.Args, arg)
	}

	return "", nil
}

const ScribeContainerPath = "/var/scribe"

func formatVolume(dir, mountPath string) string {
	return strings.Join([]string{dir, mountPath}, ":")
}

func fsArgument(dir string) (mount.Mount, error) {
	// If they've provided a directory and a separate mountpath, then we can safely not set one
	if strings.Contains(dir, ":") {
		s := strings.Split(dir, ":")
		if len(s) != 2 {
			return mount.Mount{}, errors.New("invalid format. filesystem paths should be formatted: '<source>:<target>'")
		}

		return mount.Mount{
			Source: s[0],
			Target: s[1],
		}, nil
	}

	// Relative paths should be mounted relative to /var/scribe in the container,
	// and have an absolute path for mounting (because docker).
	wd, err := os.Getwd()
	if err != nil {
		return mount.Mount{}, err
	}

	d, err := filepath.Abs(dir)
	if err != nil {
		return mount.Mount{}, err
	}

	rel, err := filepath.Rel(wd, d)
	if err != nil {
		return mount.Mount{}, err
	}

	return mount.Mount{
		Type:   mount.TypeBind,
		Source: d,
		Target: path.Join(ScribeContainerPath, rel),
	}, nil
}

// applyArguments applies a slice of arguments, typically from the requirements in a Step, onto the options
// used to run the docker container for a step.
// For example, if the step supplied requires the project (by default all of them do), then the argument type
// ArgumentTypeFS is required and is added to the RunOpts volume.
func applyArguments(configurer pipeline.Configurer, opts docker.CreateContainerOpts, args []pipeline.Argument) (docker.CreateContainerOpts, error) {
	for _, arg := range args {
		value, err := configurer.Value(arg)
		if err != nil {
			return opts, err
		}

		switch arg.Type {
		case pipeline.ArgumentTypeFS:
			mount, err := fsArgument(value)
			if err != nil {
				return opts, err
			}

			// Prefering path.Join here over filepath.Join in case any silly Windows users try to use this thing
			opts.Mounts = append(opts.Mounts, mount)
		case pipeline.ArgumentTypeString:
			// String arguments are already appended to the command and have already been placed in RunOpts; we don't need to re-implement that.
			continue
		}
	}

	return opts, nil
}
