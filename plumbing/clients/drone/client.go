package drone

import (
	"gopkg.in/yaml.v2"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/plumbing/schemas/drone"
)

var (
	ErrorNoImage = plumbing.NewPipelineError("no image provided", "An image is required for all steps in Drone. You can specify one with the '.WithImage(\"name\")' function.")
	ErrorNoName  = plumbing.NewPipelineError("no name provided", "A name is required for all steps in Drone. You can specify one with the '.WithName(\"name\")' function.")
)

type Client struct {
	Opts pipeline.CommonOpts

	Log  *plog.Logger
	List *pipeline.List
}

// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
// This function blocks the goroutine until all of the steps have completed.
func (c *Client) Run(steps ...pipeline.Step) {
	c.List.AppendLineage(steps...)
}

// Parallel will run the listed steps at the same time.
// This function blocks the goroutine until all of the steps have completed.
func (c *Client) Parallel(steps ...pipeline.Step) {
	c.List.Append(steps...)
}

func (c *Client) Validate(step pipeline.Step) error {
	if step.Image == "" {
		return ErrorNoImage
	}

	if step.Name == "" {
		return ErrorNoName
	}

	return nil
}

func (c *Client) Cache(action pipeline.StepAction, _ pipeline.Cacher) pipeline.StepAction {
	return action
}
func (c *Client) Input(_ ...pipeline.Argument) {}
func (c *Client) Output(_ ...pipeline.Output)  {}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done() {
	cfg := &drone.Pipeline{
		Name:  c.Opts.Name,
		Kind:  "pipeline",
		Type:  "docker",
		Steps: []drone.Step{},
	}

	err := c.List.Walk(func(s pipeline.Step) error {
		step, err := drone.NewStep(c, c.Opts.Args.Path, s)
		if err != nil {
			return err
		}

		cfg.Steps = append(cfg.Steps, step)
		return nil
	})

	if err != nil {
		c.Log.Fatalln("Error building drone config", err)
	}

	if err := yaml.NewEncoder(c.Opts.Output).Encode(cfg); err != nil {
		c.Log.Fatalln(err)
	}

}
