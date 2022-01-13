package shipwright

import (
	"errors"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

var (
	ErrorCLIStepHasImage = errors.New("step has a docker image specified. This may cause unexpected results if ran in CLI mode. The `-mode=docker` flag is likely more suitable")
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
		if err := ValidateCLIStep(v); err != nil {
			plog.Warnf("In step '%s': %s", v.Name, err.Error())
		}
		c.Queue.Append(v)
	}
}

func (c *CLIClient) Done() {
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

func ValidateCLIStep(step types.Step) error {
	if step.Image != "" {
		return ErrorCLIStepHasImage
	}

	return nil
}
