package docker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/docker/docker/client"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipeline/clients/cli"
	"github.com/grafana/shipwright/plumbing/pipelineutil"
	"github.com/grafana/shipwright/plumbing/plog"
	"github.com/grafana/shipwright/plumbing/stringutil"
	"github.com/grafana/shipwright/plumbing/syncutil"
	"github.com/sirupsen/logrus"
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
// Each step is ran in its own docker container.
type Client struct {
	Client client.APIClient
	Opts   pipeline.CommonOpts

	Log *logrus.Logger
}

func (c *Client) Validate(step pipeline.Step[pipeline.Action]) error {
	if step.Image == "" {
		return errors.New("no image provided")
	}
	return nil
}

// compilePipeline creates a docker container that compiles the provided pipeline so that the compiled pipeline can be mounted in
// other containers without requiring that the container has the shipwright command or go installed.
func (c *Client) compilePipeline(ctx context.Context, network *Network) (*Volume, error) {
	log := c.Log

	volume, err := CreateVolume(ctx, c.Client, CreateVolumeOpts{
		Name: fmt.Sprintf("shipwright-%s", stringutil.Random(8)),
	})

	if err != nil {
		return nil, fmt.Errorf("error creating docker volume: %w", err)
	}

	cmd := pipelineutil.GoBuild(ctx, pipelineutil.GoBuildOpts{
		Pipeline: c.Opts.Args.Path,
		Module:   "/var/shipwright",
		Output:   "/opt/shipwright/pipeline",
	})

	mounts, err := DefaultMounts(volume)
	if err != nil {
		return nil, err
	}

	opts := CreateContainerOpts{
		Name:    fmt.Sprintf("compile-%s", volume.Name),
		Image:   plumbing.SubImage("go", c.Opts.Version),
		Command: cmd.Args,
		Mounts:  mounts,
		Env: []string{
			"GOOS=linux",
			"GOARCH=amd64",
			"CGO_ENABLED=0",
		},
	}

	container, err := CreateContainer(ctx, c.Client, opts)
	if err != nil {
		return nil, err
	}

	log.Warnf("Building pipeline binary '%s' in docker volume...", c.Opts.Args.Path)
	// This should run a command very similar to this:
	// docker run --rm -v $TMPDIR:/var/shipwright shipwright/go:{version} go build -o /var/shipwright/pipeline ./{pipeline}
	if err := RunContainer(ctx, c.Client, RunContainerOpts{
		Container: container,
		Stdout:    log.WithField("stream", "stdout").Writer(),
		Stderr:    log.WithField("stream", "stderr").Writer(),
	}); err != nil {
		return nil, err
	}

	c.Log.Infof("Successfully compiled pipeline in volume '%s'", volume.Name)

	return volume, nil
}

func (c *Client) networkName() string {
	uid := stringutil.Random(8)
	return fmt.Sprintf("shipwright-%s-%s", stringutil.Slugify(c.Opts.Name), uid)
}

type walkOpts struct {
	walker  pipeline.Walker
	network *Network
	volume  *Volume
}

func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	network, err := CreateNetwork(ctx, c.Client, CreateNetworkOpts{
		Name: c.networkName(),
	})

	logger := c.Log.WithFields(plog.PipelineFields(c.Opts))
	// Every step needs a compiled version of the pipeline in order to know what to do
	// without requiring that every image has a copy of the shipwright binary
	volume, err := c.compilePipeline(ctx, network)
	if err != nil {
		return fmt.Errorf("failed to compile the pipeline in docker. Error: %w", err)
	}

	logger.Infoln("Running steps in docker")

	return c.Walk(ctx, walkOpts{
		walker:  w,
		network: network,
		volume:  volume,
	})
}

func (c *Client) stepWalkFunc(opts walkOpts) pipeline.StepWalkFunc {
	return func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
		wg := NewContainerWaitGroup()

		for _, step := range steps {
			log := c.Log.WithField("step", step.Name)

			container, err := CreateStepContainer(ctx, c.Client, CreateStepContainerOpts{
				Configurer: c,
				Step:       step,
				Network:    opts.network,
				Binary:     "/opt/shipwright/pipeline",
				Pipeline:   c.Opts.Args.Path,
				BuildID:    c.Opts.Args.BuildID,
				Volume:     opts.volume,
			})
			if err != nil {
				return err
			}

			var (
				stdout = log.WithField("stream", "stdout").Writer()
				stderr = log.WithField("stream", "stderr").Writer()
			)

			wg.Add(RunContainerOpts{
				Container: container,
				Stdout:    stdout,
				Stderr:    stderr,
			})
		}

		return wg.Wait(ctx, c.Client)
	}
}

func (c *Client) runPipeline(ctx context.Context, opts walkOpts, p pipeline.Step[pipeline.Pipeline]) error {
	return opts.walker.WalkSteps(ctx, p.Serial, c.stepWalkFunc(opts))
}

// walkPipelines returns the walkFunc that runs all of the pipelines in docker in the appropriate order.
// TODO: Most of this code looks very similar to the syncutil.StepWaitGroup and the ContainerWaitGroup type in this package.
// There should be a way to reduce it.
func (c *Client) walkPipelines(opts walkOpts) pipeline.PipelineWalkFunc {
	return func(ctx context.Context, pipelines ...pipeline.Step[pipeline.Pipeline]) error {
		var (
			wg = syncutil.NewPipelineWaitGroup()
		)

		// These pipelines run in parallel, but must all complete before continuing on to the next set.
		for i, v := range pipelines {
			// If this is a sub-pipeline, then run these steps without waiting on other pipeliens to complete.
			// However, if a sub-pipeline returns an error, then we shoud(?) stop.
			// TODO: This is probably not true... if a sub-pipeline stops then users will probably want to take note of it, but let the rest of the pipeline continue.
			// and see a report of the failure towards the end of the execution.
			if v.Type == pipeline.StepTypeSubPipeline {
				go func(ctx context.Context, p pipeline.Step[pipeline.Pipeline]) {
					if err := c.runPipeline(ctx, opts, p); err != nil {
						c.Log.WithError(err).Errorln("sub-pipeline failed")
					}
				}(ctx, pipelines[i])
				continue
			}

			// Otherwise, add this pipeline to the set that needs to complete before moving on to the next set of pipelines.
			wg.Add(pipelines[i], opts.walker, c.stepWalkFunc(opts))
		}

		if err := wg.Wait(ctx); err != nil {
			return err
		}

		return nil
	}
}

func (c *Client) Walk(ctx context.Context, opts walkOpts) error {
	return opts.walker.WalkPipelines(ctx, c.walkPipelines(opts))
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
