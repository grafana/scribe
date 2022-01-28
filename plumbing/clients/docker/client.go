package docker

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/clients/cli"
	"pkg.grafana.com/shipwright/v1/plumbing/cmdutil"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
// Each step is ran in its own docker container.
type Client struct {
	Opts  *types.CommonOpts
	Queue *types.StepQueue
}

func (c *Client) Cache(step types.StepAction, _ types.Cacher) types.StepAction {
	return step
}

func (c *Client) Validate(step types.Step) error {
	if step.Image == "" {
		return errors.New("no image provided")
	}
	return nil
}

func (c *Client) Input(_ ...types.Argument) {}
func (c *Client) Output(_ ...types.Output)  {}

// Parallel adds the list of steps into a queue to be executed concurrently
func (c *Client) Parallel(steps ...types.Step) {
	c.Queue.Append(steps...)
}

// Run adds the list of steps into a queue to be executed sequentially
func (c *Client) Run(steps ...types.Step) {
	for _, v := range steps {
		c.Queue.Append(v)
	}
}

func (c *Client) Done() {
	step := c.Opts.Args.Step
	if step != nil {
		n := *step

		for _, list := range c.Queue.Steps {
			for _, step := range list {
				if step.Serial == n {
					c.runSteps([]types.Step{step})
				}
			}
		}

		return
	}

	size := c.Queue.Size()
	i := 0
	for {
		plog.Infof("Running step(s) %d / %d", i, size)

		steps := c.Queue.Next()
		if steps == nil {
			break
		}

		if err := c.runSteps(steps); err != nil {
			plog.Fatalln("Error in step:", err)
		}

		i++
	}
}

// KnownVolumes is a map of default argument to a function used to retrieve the volume the value represents.
// For example, we know that every pipeline is ran alongisde source code.
// The user can supply a "-arg=source={path-to-source}" argument, or we can just
var KnownVolumes = map[types.StepArgument]func(*plumbing.PipelineArgs) (string, error){
	types.ArgumentSourceFS: func(args *plumbing.PipelineArgs) (string, error) {
		return ".", nil
	},
	types.ArgumentDockerSocketFS: func(*plumbing.PipelineArgs) (string, error) {
		return "/var/run/docker.sock", nil
	},
}

// GetVolumeValue will attempt to find the appropriate volume to mount based on the argument provided.
// Some arguments have known or knowable values, like "ArgumentSourceFS".
func GetVolumeValue(args *plumbing.PipelineArgs, arg types.StepArgument) (string, error) {
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
func (c *Client) Value(arg types.StepArgument) (string, error) {
	switch arg.Type {
	case types.ArgumentTypeString:
		return cli.GetArgValue(c.Opts.Args, arg)
	case types.ArgumentTypeFS:
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
func (c *Client) applyArguments(opts RunOpts, args []types.StepArgument) (RunOpts, error) {
	for _, arg := range args {
		value, err := c.Value(arg)
		if err != nil {
			return opts, err
		}

		switch arg.Type {
		case types.ArgumentTypeFS:
			volume, err := volumeValue(value)
			if err != nil {
				return opts, err
			}

			// Prefering path.Join here over filepath.Join in case any silly Windows users try to use this thing
			opts.Volumes = append(opts.Volumes, volume)
		case types.ArgumentTypeString:
			// String arguments are already appended to the command and have already been placed in RunOpts; we don't need to re-implement that.
			continue
		}
	}

	return opts, nil
}

func (c *Client) runAction(step types.Step) types.StepAction {
	cmd, err := cmdutil.StepCommand(c, c.Opts.Args.Path, step)
	if err != nil {
		plog.Fatalln(err)
		return nil
	}

	args := []string{}
	if len(cmd) > 1 {
		args = cmd[1:]
	}

	plog.Infoln(cmd[0], strings.Join(cmd[1:], " "))

	runOpts := RunOpts{
		Image:   step.Image,
		Command: cmd[0],
		Volumes: []string{},
		Args:    args,

		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	runOpts, err = c.applyArguments(runOpts, step.Arguments)
	if err != nil {
		plog.Fatalln(err)
		return nil
	}

	return func() error {
		return Run(runOpts)
	}
}

func (c *Client) runSteps(steps types.StepList) error {
	plog.Debugln("Running steps in parallel:", len(steps))
	wg := plumbing.NewWaitGroup(time.Minute)

	for _, v := range steps {
		wg.Add(c.runAction(v))
	}

	return wg.Wait()
}
