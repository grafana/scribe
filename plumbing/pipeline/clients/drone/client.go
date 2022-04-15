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

func stepsToNames[T pipeline.StepContent](steps []pipeline.Step[T]) []string {
	s := make([]string, len(steps))
	for i, v := range steps {
		s[i] = stringutil.Slugify(v.Name)
	}

	return s
}

type stepList struct {
	steps []*yaml.Container
}

func (s *stepList) AddStep(step *yaml.Container) {
	s.steps = append(s.steps, step)
}

func (c *Client) StepWalkFunc(log logrus.FieldLogger, s *stepList) func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
	return func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
		log.Debugf("Processing '%d' steps...", len(steps))
		for _, v := range steps {
			log := log.WithField("dependencies", stepsToNames(v.Dependencies))
			log.Debugf("Processing step '%s'...", v.Name)
			step, err := NewStep(c, c.Opts.Args.Path, v)
			if err != nil {
				return err
			}

			step.DependsOn = stepsToNames(v.Dependencies)
			s.AddStep(step)
			log.Debugf("Done Processing step '%s'.", v.Name)
		}
		log.Debugf("Done processing '%d' steps...", len(steps))
		return nil
	}
}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	cfg := []yaml.Resource{}
	log := c.Log.WithField("client", "drone")

	w.WalkPipelines(ctx, func(ctx context.Context, pipelines ...pipeline.Step[pipeline.Pipeline]) error {
		log.Debugf("Walking '%d' pipelines...", len(pipelines))
		for _, v := range pipelines {
			log.Debugf("Processing pipeline '%s'...", v.Name)
			sl := &stepList{}
			if err := w.WalkSteps(ctx, v.Serial, c.StepWalkFunc(log, sl)); err != nil {
				return err
			}
			pipeline := &yaml.Pipeline{
				Name:      v.Name,
				Kind:      "pipeline",
				Type:      "docker",
				Steps:     sl.steps,
				DependsOn: stepsToNames(v.Dependencies),
			}

			events := v.Content.Events
			if len(events) != 0 {
				cond, err := c.Events(events)
				if err != nil {
					return err
				}

				pipeline.Trigger = cond
			}

			log.Debugf("Done processing pipeline '%s'", v.Name)
			cfg = append(cfg, pipeline)
		}
		return nil
	})

	manifest := &yaml.Manifest{
		Resources: cfg,
	}

	pretty.Print(c.Opts.Output, manifest)
	return nil
}
