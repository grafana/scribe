package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
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
	Log  *logrus.Logger
}

func (c *Client) Validate(step pipeline.Step) error {
	if step.Image != "" {
		c.Log.Debugln(fmt.Sprintf("[%s]", step.Name), ErrorCLIStepHasImage.Error())
	}

	return nil
}

func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	logWrapper := &wrappers.LogWrapper{
		Opts: c.Opts,
		Log:  c.Log,
	}

	traceWrapper := &wrappers.TraceWrapper{
		Opts:   c.Opts,
		Tracer: c.Opts.Tracer,
	}

	// Because these wrappers wrap the actions of each step, the first wrapper typically runs first.
	walkFunc := traceWrapper.Wrap(c.WalkFunc)
	walkFunc = logWrapper.Wrap(walkFunc)

	return w.Walk(ctx, walkFunc)
}

func (c *Client) runSteps(ctx context.Context, steps pipeline.StepList) error {
	c.Log.Debugln("Running steps in parallel:", len(steps))

	var (
		wg   = syncutil.NewWaitGroup()
		opts = pipeline.ActionOpts{}
	)

	for _, v := range steps {
		wg.Add(v)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	return wg.Wait(ctx, opts)
}
