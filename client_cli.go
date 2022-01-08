package shipwright

import (
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// The CLIClient is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
type CLIClient struct {
	Opts  *CommonOpts
	Queue *types.StepQueue
}

func (c *CLIClient) Cache(step types.StepAction, _ types.Cacher) types.StepAction {
	return step
}

func (c *CLIClient) Input(_ ...Argument) {}
func (c *CLIClient) Output(_ ...Output)  {}

func (c *CLIClient) Init(opts *CommonOpts) {
	c.Opts = opts
}

// Parallel adds the list of steps into a queue to be executed concurrently
func (c *CLIClient) Parallel(steps ...types.Step) {
	c.Queue.Append(steps...)
}

// Run adds the list of steps into a queue to be executed sequentially
func (c *CLIClient) Run(steps ...types.Step) {
	for _, v := range steps {
		c.Queue.Append(v)
	}
}

func (c *CLIClient) Done() {
	step := c.Opts.Args.Step
	if step != nil {
		n := *step

		if err := c.runSteps(c.Queue.At(n)); err != nil {
			plog.Fatalln("Error in step:", err)
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

func NewCLIClient(opts *CommonOpts) Shipwright {
	return Shipwright{
		Client: &CLIClient{
			Opts:  opts,
			Queue: &types.StepQueue{},
		},
	}
}

func (c *CLIClient) runSteps(steps types.StepList) error {
	plog.Debugln("Running steps in parallel:", len(steps))
	wg := plumbing.NewWaitGroup(time.Minute)

	for _, v := range steps {
		wg.Add(v.Action)
	}

	return wg.Wait()
}
