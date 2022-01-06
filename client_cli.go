package shipwright

import (
	"log"
	"time"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/cmd/commands"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// The CLIClient is used when interacting with a shipwright pipeline using the shipwright CLI.
// In order to emulate what happens in a remote environment, the steps are put into a queue before being ran.
type CLIClient struct {
	Opts  *commands.RunArgs
	Queue *types.StepQueue
}

func (c *CLIClient) Cache(step types.Step, _ types.Cacher) types.Step {
	return step
}

func (c *CLIClient) Input(_ ...Argument) {}
func (c *CLIClient) Output(_ ...Output)  {}

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
	//
	if c.Opts.Step != nil {
		n := *c.Opts.Step

		if err := c.runSteps(c.Queue.At(n)); err != nil {
			log.Fatalln("Error in step:", err)
		}

		return
	}

	size := c.Queue.Size()
	i := 0
	for {
		log.Printf("Running step(s) %d / %d", i, size)

		steps := c.Queue.Next()
		if steps == nil {
			break
		}

		if err := c.runSteps(steps); err != nil {
			log.Fatalln("Error in step:", err)
		}

		i++
	}
}

// The CLIClient uses the same arguments as the "shipwright run" command, and so it calls that function rather than handling those values here.
func (c *CLIClient) Parse(args []string) error {
	opts, err := commands.ParseRunArgs(args)
	if err != nil {
		return err
	}

	c.Opts = opts

	return nil
}

func NewCLIClient() Shipwright {
	return Shipwright{
		Client: &CLIClient{
			Queue: &types.StepQueue{},
		},
	}
}

func (c *CLIClient) runSteps(steps types.StepList) error {
	log.Println("Running steps in parallel:", len(steps))
	wg := plumbing.NewWaitGroup(time.Minute)

	for _, v := range steps {
		wg.Add(v)
	}

	return wg.Wait()
}
