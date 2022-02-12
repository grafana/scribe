package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/syncutil"
)

var (
	ErrorCLIStepHasImage = errors.New("step has a docker image specified. This may cause unexpected results if ran in CLI mode. The `-mode=docker` flag is likely more suitable")
)

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
type Client struct {
	Opts pipeline.CommonOpts

	Log   *plog.Logger
	Queue *pipeline.StepQueue
}

func (c *Client) Cache(step pipeline.StepAction, _ pipeline.Cacher) pipeline.StepAction {
	return step
}

func (c *Client) Input(_ ...pipeline.Argument) {}
func (c *Client) Output(_ ...pipeline.Output)  {}

func (c *Client) Validate(step pipeline.Step) error {
	if step.Image != "" {
		c.Log.Debugln(ErrorCLIStepHasImage.Error())
	}

	return nil
}

// Parallel adds the list of steps into a queue to be executed concurrently
func (c *Client) Parallel(steps ...pipeline.Step) {
	c.Queue.Append(steps...)
}

// Run adds the list of steps into a queue to be executed sequentially
func (c *Client) Run(steps ...pipeline.Step) {
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
					c.runSteps(ctx, []pipeline.Step{step})
				}
			}
		}

		return
	}

	size := c.Queue.Size()
	i := 0
	for {
		steps := c.Queue.Next()
		if steps == nil {
			break
		}

		c.Log.Infof("Running step(s) %d / %d %s", i, size, steps.String())

		if err := c.runSteps(ctx, steps); err != nil {
			c.Log.Fatalln("Error in step:", err)
		}

		i++
	}
}

func (c *Client) wrap(step pipeline.Step) pipeline.Step {
	action := step.Action
	step.Action = func(opts pipeline.ActionOpts) error {
		var (
			stdout = bytes.NewBuffer(nil)
			stderr = bytes.NewBuffer(nil)
		)
		opts.Stdout = stdout
		opts.Stderr = stderr
		if err := action(opts); err != nil {
			return fmt.Errorf("error: %w\nstdout:%s\nstderr:%s", err, stdout.String(), stderr.String())
		}

		return nil
	}

	return step
}

func (c *Client) runSteps(ctx context.Context, steps pipeline.StepList) error {
	c.Log.Debugln("Running steps in parallel:", len(steps))
	var (
		wg   = syncutil.NewWaitGroup()
		opts = pipeline.ActionOpts{}
	)

	for _, v := range steps {
		wg.Add(c.wrap(v))
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	return wg.Wait(ctx, opts)
}
