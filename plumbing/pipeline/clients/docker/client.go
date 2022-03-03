package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	golangx "pkg.grafana.com/shipwright/v1/golang/x"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/cmdutil"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/syncutil"
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
// Each step is ran in its own docker container.
type Client struct {
	Opts pipeline.CommonOpts

	Log *plog.Logger
}

func (c *Client) Validate(step pipeline.Step) error {
	if step.Image == "" {
		return errors.New("no image provided")
	}
	return nil
}

// buildPipeline compiles the provided pipeline so that it can be mounted in a container without requiring that the pipeline has the shipwright command or go installed.
func (c *Client) buildPipeline(ctx context.Context) (string, error) {
	p, err := os.MkdirTemp(os.TempDir(), "shipwright-")
	if err != nil {
		return "", err
	}

	path := filepath.Join(p, "pipeline")
	var (
		stdout = bytes.NewBuffer(nil)
		stderr = bytes.NewBuffer(nil)
	)

	c.Log.Warnf("Building pipeline binary '%s' at '%s' for use in docker container...", c.Opts.Args.Path, path)
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if err := golangx.Build(ctx, golangx.BuildOpts{
		Pkg:    c.Opts.Args.Path,
		Module: wd,
		Output: path,
		Stdout: stdout,
		Stderr: stderr,
	}); err != nil {
		return "", fmt.Errorf("error: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	return path, nil
}

func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	// Every step needs a compiled version of the pipeline in order to know what to do
	// without requiring that every image has a copy of the shipwright binary
	p, err := c.buildPipeline(ctx)
	if err != nil {
		c.Log.Fatalln("failed to compile the pipeline", err.Error())
	}

	return w.Walk(ctx, func(ctx context.Context, steps ...pipeline.Step) error {
		s := make([]string, len(steps))
		for i, v := range steps {
			s[i] = v.Name
		}

		c.Log.Infof("Running [%d] step(s) %s", len(steps), len(steps), strings.Join(s, " | "))

		if err := c.runSteps(ctx, p, steps); err != nil {
			return err
		}

		return nil
	})
}

// KnownVolumes is a map of default argument to a function used to retrieve the volume the value represents.
// For example, we know that every pipeline is ran alongisde source code.
// The user can supply a "-arg=source={path-to-source}" argument, or we can just
var KnownVolumes = map[pipeline.Argument]func(*plumbing.PipelineArgs) (string, error){
	pipeline.ArgumentSourceFS: func(args *plumbing.PipelineArgs) (string, error) {
		return ".", nil
	},
	pipeline.ArgumentDockerSocketFS: func(*plumbing.PipelineArgs) (string, error) {
		return "/var/run/docker.sock", nil
	},
}

// GetVolumeValue will attempt to find the appropriate volume to mount based on the argument provided.
// Some arguments have known or knowable values, like "ArgumentSourceFS".
func GetVolumeValue(args *plumbing.PipelineArgs, arg pipeline.Argument) (string, error) {
	// If an applicable argument is provided, then we should use that, even if it's a known value.
	if val, err := args.ArgMap.Get(arg.Key); err == nil {
		return val, nil
	}

	// See if we can find a known value for this FS...
	if f, ok := KnownVolumes[arg]; ok {
		return f(args)
	}

	// TODO: Should we request via stdin?
	return "", nil
}

// Value retrieves the configuration item the same way the CLI does; by looking in the argmap or by asking via stdin.
func (c *Client) Value(arg pipeline.Argument) (string, error) {
	switch arg.Type {
	case pipeline.ArgumentTypeString:
		return cli.GetArgValue(c.Opts.Args, arg)
	case pipeline.ArgumentTypeFS:
		return GetVolumeValue(c.Opts.Args, arg)
	}

	return "", nil
}

const ShipwrightContainerPath = "/var/shipwright"

func formatVolume(dir, mountPath string) string {
	return strings.Join([]string{dir, mountPath}, ":")
}

func volumeValue(dir string) (string, error) {
	// If they've provided a directory and a separate mountpath, then we can safely not set one
	if strings.Contains(dir, ":") {
		return dir, nil
	}

	// Absolute paths should be preserved
	if filepath.IsAbs(dir) {
		return formatVolume(dir, dir), nil
	}

	// Relative paths should be mounted relative to /var/shipwright in the container,
	// and have an absolute path for mounting (because docker).

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	d, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(wd, d)
	if err != nil {
		return "", err
	}

	return formatVolume(d, path.Join(ShipwrightContainerPath, rel)), nil
}

// applyArguments applies a slice of arguments, typically from the requirements in a Step, onto the options
// used to run the docker container for a step.
// For example, if the step supplied requires the project (by default all of them do), then the argument type
// ArgumentTypeFS is required and is added to the RunOpts volume.
func (c *Client) applyArguments(opts RunOpts, args []pipeline.Argument) (RunOpts, error) {
	for _, arg := range args {
		value, err := c.Value(arg)
		if err != nil {
			return opts, err
		}

		switch arg.Type {
		case pipeline.ArgumentTypeFS:
			volume, err := volumeValue(value)
			if err != nil {
				return opts, err
			}

			// Prefering path.Join here over filepath.Join in case any silly Windows users try to use this thing
			opts.Volumes = append(opts.Volumes, volume)
		case pipeline.ArgumentTypeString:
			// String arguments are already appended to the command and have already been placed in RunOpts; we don't need to re-implement that.
			continue
		}
	}

	return opts, nil
}

func (c *Client) runAction(ctx context.Context, pipelinePath string, step pipeline.Step) pipeline.StepAction {
	cmd, err := cmdutil.StepCommand(c, "", step)
	if err != nil {
		c.Log.Fatalln(err)
		return nil
	}

	args := []string{}
	if len(cmd) > 1 {
		args = cmd[1:]
	}

	runOpts := RunOpts{
		PipelinePath: pipelinePath,
		Image:        step.Image,
		Command:      PipelineVolumePath,
		Volumes:      []string{},
		Args:         args,
	}

	runOpts, err = c.applyArguments(runOpts, step.Arguments)
	if err != nil {
		c.Log.Fatalln(err)
		return nil
	}

	return func(ctx context.Context, opts pipeline.ActionOpts) error {
		runOpts.Stdout = opts.Stdout
		runOpts.Stderr = opts.Stderr

		c.Log.Debugf("Running command 'docker %v'", RunArgs(runOpts))
		return Run(ctx, runOpts)
	}
}

func (c *Client) wrap(pipelinePath string, step pipeline.Step) pipeline.Step {
	step.Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
		opts.Stdout = os.Stdout
		opts.Stderr = os.Stderr

		if err := c.runAction(ctx, pipelinePath, step)(ctx, opts); err != nil {
			return err
		}

		return nil
	}

	return step
}

func (c *Client) runSteps(ctx context.Context, pipelinePath string, steps pipeline.StepList) error {
	c.Log.Debugln("Running steps in parallel:", len(steps))

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	wg := syncutil.NewWaitGroup()

	opts := pipeline.ActionOpts{}
	for _, v := range steps {
		wg.Add(c.wrap(pipelinePath, v))
	}

	return wg.Wait(ctx, opts)
}
