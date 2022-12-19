package drone

import (
	"context"
	"net/url"
	"path"

	"github.com/drone/drone-yaml/yaml"
	"github.com/drone/drone-yaml/yaml/pretty"
	"github.com/grafana/scribe/errors"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/pipeline/clients"
	"github.com/grafana/scribe/pipelineutil"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/stringutil"
	"github.com/sirupsen/logrus"
)

var (
	ErrorNoImage = errors.NewPipelineError("no image provided", "An image is required for all steps in Drone. You can specify one with the '.WithImage(\"name\")' function.")
	ErrorNoName  = errors.NewPipelineError("no name provided", "A name is required for all steps in Drone. You can specify one with the '.WithName(\"name\")' function.")
)

// Client is the Drone implementation of the pipeline Client interface.
// It will create a `.drone.yml` file that will run your pipeline the same way that it would run locally.
// It uses the dagger client to run an entire pipeline as a single step. While not ideal for visualization purposes, this does behavior is what enables consistency.
type Client struct {
	Opts clients.CommonOpts

	Log *logrus.Logger
}

// Validate ensures that your step has a name and a docker image.
func (c *Client) Validate(step pipeline.Step) error {
	if step.Image == "" {
		return ErrorNoImage
	}

	if step.Name == "" {
		return ErrorNoName
	}

	return nil
}

func stepsToNames(steps []pipeline.Step) []string {
	s := make([]string, len(steps))
	for i, v := range steps {
		s[i] = stringutil.Slugify(v.Name)
	}

	return s
}

func pipelinesToNames(steps []pipeline.Pipeline) []string {
	s := make([]string, len(steps))
	for i, v := range steps {
		s[i] = stringutil.Slugify(v.Name)
	}

	return s
}

type stepList struct {
	steps    []*yaml.Container
	services []*yaml.Container
}

func (s *stepList) AddStep(step *yaml.Container) {
	s.steps = append(s.steps, step)
}

func (s *stepList) AddService(step *yaml.Container) {
	s.services = append(s.services, step)
}

func (c *Client) Step(v pipeline.Pipeline, state string) (*yaml.Container, error) {
	step, err := NewDaggerStep(c, c.Opts.Args.Path, state, c.Opts.Version, v)
	if err != nil {
		return nil, err
	}
	return step, nil
}

var (
	PipelinePath = "/var/scribe/pipeline"
	StatePath    = "/var/scribe-state"
	ScribeVolume = &yaml.Volume{
		Name:     "scribe",
		EmptyDir: &yaml.VolumeEmptyDir{},
	}
	HostDockerVolume = &yaml.Volume{
		Name: stringutil.Slugify(pipeline.ArgumentDockerSocketFS.Key),
		HostPath: &yaml.VolumeHostPath{
			Path: "/var/run/docker.sock",
		},
	}
	ScribeStateVolume = &yaml.Volume{
		Name:     "scribe-state",
		EmptyDir: &yaml.VolumeEmptyDir{},
	}
	ScribeVolumeMount = &yaml.VolumeMount{
		Name:      "scribe",
		MountPath: "/var/scribe",
	}
	ScribeStateVolumeMount = &yaml.VolumeMount{
		Name:      "scribe-state",
		MountPath: StatePath,
	}
)

type newPipelineOpts struct {
	Name      string
	Steps     []*yaml.Container
	Services  []*yaml.Container
	DependsOn []string
}

func (c *Client) newPipeline(opts newPipelineOpts, pipelineOpts clients.CommonOpts) *yaml.Pipeline {
	command := pipelineutil.GoBuild(context.Background(), pipelineutil.GoBuildOpts{
		Pipeline: pipelineOpts.Args.Path,
		Output:   PipelinePath,
	})

	build := &yaml.Container{
		Name:    "builtin-compile-pipeline",
		Image:   "golang:1.19",
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
		Volumes: []*yaml.VolumeMount{ScribeVolumeMount},
	}

	// Add the approprirate steps and "DependsOn" to each step.
	for i, v := range opts.Steps {
		if len(v.DependsOn) == 0 {
			opts.Steps[i].DependsOn = []string{build.Name}
		}
		opts.Steps[i].Volumes = append(opts.Steps[i].Volumes, ScribeVolumeMount, ScribeStateVolumeMount)
	}

	p := &yaml.Pipeline{
		Name:      opts.Name,
		Kind:      "pipeline",
		Type:      "docker",
		DependsOn: opts.DependsOn,
		Steps:     append([]*yaml.Container{build}, opts.Steps...),
		Services:  opts.Services,
		Volumes: []*yaml.Volume{
			ScribeVolume,
			ScribeStateVolume,
			HostDockerVolume,
		},
	}

	return p
}

// Done traverses through the tree and writes a .drone.yml file to the provided writer
func (c *Client) Done(ctx context.Context, w *pipeline.Collection) error {
	cfg := []yaml.Resource{}
	log := c.Log.WithField("client", "drone")

	// StatePath is an aboslute path and already has a '/'.
	stateArg := &url.URL{
		Scheme: "file",
		Path:   path.Join(StatePath, "state.json"),
	}

	pipelines := []pipeline.Pipeline{}

	if err := w.WalkPipelines(ctx, func(ctx context.Context, p pipeline.Pipeline) error {
		pipelines = append(pipelines, p)
		return nil
	}); err != nil {
		return err
	}

	for _, v := range pipelines {
		if v.ID == 0 {
			continue
		}
		log.Debugf("Processing pipeline '%s'...", v.Name)

		s, err := c.Step(v, stateArg.String())
		if err != nil {
			return err
		}

		dependencies := []string{}
		// Find the pipelines that supply the argument that we require.
		// The collection type has already verified that everything we are expecting exists somewhere in the graph.
		for _, arg := range v.RequiredArgs {
			for _, p := range pipelines {
				if state.ArgListContains(p.ProvidedArgs, arg) {
					dependencies = append(dependencies, stringutil.Slugify(p.Name))
					break
				}
			}
		}

		pipeline := c.newPipeline(newPipelineOpts{
			Name:      stringutil.Slugify(v.Name),
			Steps:     []*yaml.Container{s},
			DependsOn: dependencies,
		}, c.Opts)
		if len(v.Events) == 0 {
			log.Debugf("Pipeline '%d' / '%s' has 0 events", v.ID, v.Name)
		} else {
			log.Debugf("Pipeline '%d' / '%s' has '%d events", v.ID, v.Name, len(v.Events))
		}
		events := v.Events
		if len(events) != 0 {
			log.Debugf("Generating with %d event filters...", len(events))
			cond, err := Events(events)
			if err != nil {
				return err
			}

			pipeline.Trigger = cond
		}

		log.Debugf("Done processing pipeline '%s'", v.Name)
		cfg = append(cfg, pipeline)
	}

	manifest := &yaml.Manifest{
		Resources: cfg,
	}
	pretty.Print(c.Opts.Output, manifest)

	return nil
}
