package drone

import (
	"context"

	"github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone-yaml/yaml/pretty"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/stringutil"
	"github.com/sirupsen/logrus"
)

var (
	ErrorNoImage = plumbing.NewPipelineError("no image provided", "An image is required for all steps in Drone. You can specify one with the '.WithImage(\"name\")' function.")
	ErrorNoName  = plumbing.NewPipelineError("no name provided", "A name is required for all steps in Drone. You can specify one with the '.WithName(\"name\")' function.")
)

type Client struct {
	Opts pipeline.CommonOpts

	Log *logrus.Logger
}

func (c *Client) Validate(step pipeline.Step[pipeline.Action]) error {
	if step.Image == "" {
		return ErrorNoImage
	}

	if step.Name == "" {
		return ErrorNoName
	}

	return nil
}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done(ctx context.Context, w pipeline.Walker, events []pipeline.Event) error {
	cfg := &yaml.Pipeline{
		Name:  c.Opts.Name,
		Kind:  "pipeline",
		Type:  "docker",
		Steps: []*yaml.Container{},
	}

	if len(events) != 0 {
		cond, err := c.Events(events)
		if err != nil {
			return err
		}

		cfg.Trigger = cond
	}

	previous := []string{}

	// When walking through each list of steps, we assume that the previous list of steps are required before this one will run.
	// It's entirely possible in the future, when this Walk function is backed by a DAG, we can't safely make that assumption. Instead, we will have to defer to the parent nodes and use those as "DependsOn"
	if err := w.WalkSteps(ctx, 1, func(ctx context.Context, s ...pipeline.Step[pipeline.Action]) error {
		stepNames := make([]string, len(s))
		for i, v := range s {
			step, err := NewStep(c, c.Opts.Args.Path, v)
			if err != nil {
				return err
			}

			stepNames[i] = stringutil.Slugify(v.Name)

			step.DependsOn = previous
			cfg.Steps = append(cfg.Steps, step)
		}

		previous = stepNames

		return nil
	}); err != nil {
		return err
	}

	manifest := &yaml.Manifest{
		Resources: []yaml.Resource{cfg},
	}

	pretty.Print(c.Opts.Output, manifest)
	return nil
}
