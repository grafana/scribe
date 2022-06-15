package docker

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/grafana/scribe/plumbing/pipeline"
)

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

func fsArgument(dir string) (docker.HostMount, error) {
	// If they've provided a directory and a separate mountpath, then we can safely not set one
	if strings.Contains(dir, ":") {
		s := strings.Split(dir, ":")
		if len(s) != 2 {
			return docker.HostMount{}, errors.New("invalid format. filesystem paths should be formatted: '<source>:<target>'")
		}

		return docker.HostMount{
			Type:   "bind",
			Source: s[0],
			Target: s[1],
		}, nil
	}

	// Relative paths should be mounted relative to /var/scribe in the container,
	// and have an absolute path for mounting (because docker).
	wd, err := os.Getwd()
	if err != nil {
		return docker.HostMount{}, err
	}

	d, err := filepath.Abs(dir)
	if err != nil {
		return docker.HostMount{}, err
	}

	rel, err := filepath.Rel(wd, d)
	if err != nil {
		return docker.HostMount{}, err
	}

	return docker.HostMount{
		Type:   "bind",
		Source: d,
		Target: path.Join(ScribeContainerPath, rel),
	}, nil
}

// applyArguments applies a slice of arguments, typically from the requirements in a Step, onto the options
// used to run the docker container for a step.
// For example, if the step supplied requires the project (by default all of them do), then the argument type
// ArgumentTypeFS is required and is added to the RunOpts volume.
func applyArguments(configurer pipeline.Configurer, opts docker.CreateContainerOptions, args []pipeline.Argument) (docker.CreateContainerOptions, error) {
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
			opts.HostConfig.Mounts = append(opts.HostConfig.Mounts, mount)
		case pipeline.ArgumentTypeString:
			// String arguments are already appended to the command and have already been placed in RunOpts; we don't need to re-implement that.
			continue
		}
	}

	return opts, nil
}
