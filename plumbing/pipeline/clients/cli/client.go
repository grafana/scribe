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
	"pkg.grafana.com/shipwright/v1/plumbing/wrappers"
)

var (
	ErrorCLIStepHasImage = errors.New("step has a docker image specified. This may cause unexpected results if ran in CLI mode. The `-mode=docker` flag is likely more suitable")
)

// WalkFunc walks through the steps that the collector provides
func (c *Client) WalkFunc(ctx context.Context, step ...pipeline.Step) error {
	if err := c.runSteps(ctx, step); err != nil {
		return err
	}

	return nil
}

// The Client is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
type Client struct {
	Opts pipeline.CommonOpts
	Log  *plog.Logger
}

func (c *Client) Validate(step pipeline.Step) error {
	if step.Image != "" {
		c.Log.Debugln(fmt.Sprintf("[%s]", step.Name), ErrorCLIStepHasImage.Error())
	}

	return nil
}

func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	wrapper := &wrappers.LogWrapper{
		Log: c.Log,
	}

	return w.Walk(ctx, wrapper.Wrap(c.WalkFunc))
}

func (c *Client) wrap(step pipeline.Step) pipeline.Step {
	action := step.Action
	step.Action = func(ctx context.Context, opts pipeline.ActionOpts) error {
		var (
			stdout = bytes.NewBuffer(nil)
			stderr = bytes.NewBuffer(nil)
		)

		opts.Stdout = stdout
		opts.Stderr = stderr

		if err := action(ctx, opts); err != nil {
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