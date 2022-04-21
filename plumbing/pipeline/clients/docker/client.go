package docker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipelineutil"
	"github.com/grafana/shipwright/plumbing/plog"
	"github.com/grafana/shipwright/plumbing/syncutil"
	"github.com/sirupsen/logrus"
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
// Each step is ran in its own docker container.
type Client struct {
	Opts pipeline.CommonOpts

	Log *logrus.Logger
}

func (c *Client) Validate(step pipeline.Step[pipeline.Action]) error {
	if step.Image == "" {
		return errors.New("no image provided")
	}
	return nil
}

// buildPipeline creates a docker container that compiles the provided pipeline so that the compiled pipeline can be mounted in
// other containers without requiring that the container has the shipwright command or go installed.
func (c *Client) buildPipeline(ctx context.Context) (string, error) {
	p, err := os.MkdirTemp(os.TempDir(), "shipwright-")
	if err != nil {
		return "", fmt.Errorf("error creating temporary directory: %w", err)
	}

	c.Log.Warnf("Building pipeline binary '%s' at '%s' for use in docker container...", c.Opts.Args.Path, p)
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}

	env := []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	}

	output := filepath.Join(p, "pipeline")
	cmd := pipelineutil.GoBuild(ctx, pipelineutil.GoBuildOpts{
		Pipeline: c.Opts.Args.Path,
		Module:   wd,
		Output:   output,
	})

	opts := RunOpts{
		Stdout:  c.Log.WithField("stream", "stdout").Writer(),
		Stderr:  c.Log.WithField("stream", "stderr").Writer(),
		Image:   plumbing.SubImage("go", c.Opts.Version),
		Command: cmd.Args[0],
		Args:    cmd.Args[1:],
		Env:     env,
		Volumes: []string{
			fmt.Sprintf("%s:%s", p, p),
			fmt.Sprintf("%s:/var/shipwright", wd),
		},
	}

	c.Log.Infof("Running docker command '%s'", append([]string{"docker"}, RunArgs(opts)...))
	// This should run a command very similar to this:
	// docker run --rm -v $TMPDIR:/var/shipwright shipwright/go:{version} go build -o /var/shipwright/pipeline ./{pipeline}
	if err := Run(ctx, opts); err != nil {
		return "", err
	}

	c.Log.Infof("Successfully compiled pipeline at '%s'", output)

	return output, nil
}

func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	logger := c.Log.WithFields(plog.PipelineFields(c.Opts))

	// Every step needs a compiled version of the pipeline in order to know what to do
	// without requiring that every image has a copy of the shipwright binary
	_, err := c.buildPipeline(ctx)
	if err != nil {
		return fmt.Errorf("failed to compile the pipeline in docker. Error: %w", err)
	}

	// return w.WalkSteps(ctx, 0, func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {})
	return nil
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

func (c *Client) runAction(ctx context.Context, pipelinePath string, step pipeline.Step[pipeline.Action]) pipeline.Action {
	cmd, err := cmdutil.StepCommand(c, cmdutil.CommandOpts{
		Step:    step,
		BuildID: c.Opts.Args.BuildID,
	})
	if err != nil {
		c.Log.Fatalln(err)
		return nil
	}

	args := []string{}
	if len(cmd) > 1 {
		args = cmd[1:]
	}

	runOpts := RunOpts{
		Image:   step.Image,
		Command: PipelineVolumePath,
		Volumes: []string{},
		Args:    args,
	}

	runOpts = runOpts.WithPipelinePath(pipelinePath)

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

func (c *Client) wrap(pipelinePath string, step pipeline.Step[pipeline.Action]) pipeline.Step[pipeline.Action] {
	step.Content = func(ctx context.Context, opts pipeline.ActionOpts) error {
		opts.Stdout = c.Log.WithField("stream", "stdout").Writer()
		opts.Stderr = c.Log.WithField("stream", "stderr").Writer()

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
