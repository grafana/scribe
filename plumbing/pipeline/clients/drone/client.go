package drone

import (
	"context"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/schemas/drone"
)

var (
	ErrorNoImage = plumbing.NewPipelineError("no image provided", "An image is required for all steps in Drone. You can specify one with the '.WithImage(\"name\")' function.")
	ErrorNoName  = plumbing.NewPipelineError("no name provided", "A name is required for all steps in Drone. You can specify one with the '.WithName(\"name\")' function.")
)

type Client struct {
	Opts pipeline.CommonOpts

	Log *logrus.Logger
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

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	cfg := &drone.Pipeline{
		Name:  c.Opts.Name,
		Kind:  "pipeline",
		Type:  "docker",
		Steps: []drone.Step{},
	}

	previous := []string{}

	// When walking through each list of steps, we assume that the previous list of steps are required before this one will run.
	// It's entirely possible in the future, when this Walk function is backed by a DAG, we can't safely make that assumption. Instead, we will have to defer to the parent nodes and use those as "DependsOn"
	err := w.Walk(ctx, func(ctx context.Context, s ...pipeline.Step) error {
		stepNames := make([]string, len(s))
		for i, v := range s {
			step, err := drone.NewStep(c, c.Opts.Args.Path, v)
			if err != nil {
				return err
			}

			stepNames[i] = drone.Slugify(v.Name)

			step.DependsOn = previous
			cfg.Steps = append(cfg.Steps, step)
		}

		previous = stepNames

		return nil
	})

	if err != nil {
		return err
	}

	if err := yaml.NewEncoder(c.Opts.Output).Encode(cfg); err != nil {
		return err
	}

	return nil
}
