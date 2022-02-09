package cli

import (
	"context"
	"errors"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/syncutil"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

var (
	ErrorCLIStepHasImage = errors.New("step has a docker image specified. This may cause unexpected results if ran in CLI mode. The `-mode=docker` flag is likely more suitable")
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
type Client struct {
	Opts  *types.CommonOpts
	Queue *types.StepQueue
}

func (c *Client) Cache(step types.StepAction, _ types.Cacher) types.StepAction {
	return step
}

func (c *Client) Input(_ ...types.Argument) {}
func (c *Client) Output(_ ...types.Output)  {}

func (c *Client) Validate(step types.Step) error {
	if step.Image != "" {
		plog.Debugln(ErrorCLIStepHasImage.Error())
	}

	return nil
}

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
	ctx := context.Background()

	step := c.Opts.Args.Step
	if step != nil {
		n := *step

		for _, list := range c.Queue.Steps {
			for _, step := range list {
				if step.Serial == n {
					c.runSteps(ctx, []types.Step{step})
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

		if err := c.runSteps(ctx, steps); err != nil {
			plog.Fatalln("Error in step:", err)
		}

		i++
	}
}

func (c *Client) wrap(step types.Step) types.Step {
	action := step.Action
	step.Action = func(opts types.ActionOpts) error {
		return action(opts)
	}

	return step
}

func (c *Client) runSteps(ctx context.Context, steps types.StepList) error {
	plog.Debugln("Running steps in parallel:", len(steps))
	var (
		wg   = syncutil.NewWaitGroup(time.Minute)
		opts = types.ActionOpts{}
	)

	for _, v := range steps {
		wg.Add(c.wrap(v))
	}

	return wg.Wait(ctx, opts)
}
