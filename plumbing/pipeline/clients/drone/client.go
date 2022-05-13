package drone

import (
	"context"
	"net/url"
	"path"

	"github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone-yaml/yaml/pretty"
	"github.com/grafana/shipwright/plumbing"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/pipelineutil"
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

func (c *Client) StepWalkFunc(log logrus.FieldLogger, s *stepList, state string) func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
	return func(ctx context.Context, steps ...pipeline.Step[pipeline.Action]) error {
		log.Debugf("Processing '%d' steps...", len(steps))
		for _, v := range steps {
			log := log.WithField("dependencies", stepsToNames(v.Dependencies))
			log.Debugf("Processing step '%s'...", v.Name)
			step, err := NewStep(c, c.Opts.Args.Path, state, v)
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

var (
	PipelinePath     = "/var/shipwright/pipeline"
	StatePath        = "/var/shipwright-state"
	ShipwrightVolume = &yaml.Volume{
		Name:     "shipwright",
		EmptyDir: &yaml.VolumeEmptyDir{},
	}
	ShipwrightStateVolume = &yaml.Volume{
		Name:     "shipwright-state",
		EmptyDir: &yaml.VolumeEmptyDir{},
	}
	ShipwrightVolumeMount = &yaml.VolumeMount{
		Name:      "shipwright",
		MountPath: "/var/shipwright",
	}
	ShipwrightStateVolumeMount = &yaml.VolumeMount{
		Name:      "shipwright-state",
		MountPath: StatePath,
	}
)

type newPipelineOpts struct {
	Name      string
	Steps     []*yaml.Container
	DependsOn []string
}

func (c *Client) newPipeline(opts newPipelineOpts, pipelineOpts pipeline.CommonOpts) *yaml.Pipeline {
	command := pipelineutil.GoBuild(context.Background(), pipelineutil.GoBuildOpts{
		Pipeline: pipelineOpts.Args.Path,
		Output:   PipelinePath,
	})

	build := &yaml.Container{
		Name:    "builtin-compile-pipeline",
		Image:   plumbing.SubImage("go", c.Opts.Version),
		Command: command.Args,
		Environment: map[string]*yaml.Variable{
			"GOOS": {
				Value: "linux",
			},
			"GOARCH": {
				Value: "amd64",
			},
			"CGO_ENABLED": {
				Value: "0",
			},
		},
		Volumes: []*yaml.VolumeMount{ShipwrightVolumeMount},
	}

	for i, v := range opts.Steps {
		if len(v.DependsOn) == 0 {
			opts.Steps[i].DependsOn = []string{build.Name}
		}
		opts.Steps[i].Volumes = append(opts.Steps[i].Volumes, ShipwrightVolumeMount, ShipwrightStateVolumeMount)
	}

	p := &yaml.Pipeline{
		Name:      opts.Name,
		Kind:      "pipeline",
		Type:      "docker",
		DependsOn: opts.DependsOn,
		Steps:     append([]*yaml.Container{build}, opts.Steps...),
		Volumes: []*yaml.Volume{
			ShipwrightVolume,
			ShipwrightStateVolume,
		},
	}

	return p
}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done(ctx context.Context, w pipeline.Walker) error {
	cfg := []yaml.Resource{}
	log := c.Log.WithField("client", "drone")

	// StatePath is an aboslute path and already has a '/'.
	state := &url.URL{
		Scheme: "file",
		Path:   path.Join(StatePath, "state.json"),
	}

	w.WalkPipelines(ctx, func(ctx context.Context, pipelines ...pipeline.Step[pipeline.Pipeline]) error {
		log.Debugf("Walking '%d' pipelines...", len(pipelines))
		for _, v := range pipelines {
			log.Debugf("Processing pipeline '%s'...", v.Name)
			sl := &stepList{}
			if err := w.WalkSteps(ctx, v.Serial, c.StepWalkFunc(log, sl, state.String())); err != nil {
				return err
			}

			pipeline := c.newPipeline(newPipelineOpts{
				Name:      v.Name,
				Steps:     sl.steps,
				DependsOn: stepsToNames(v.Dependencies),
			}, c.Opts)

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
