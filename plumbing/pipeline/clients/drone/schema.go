package drone

import (
	"fmt"
	"strings"

	"github.com/drone/drone-yaml/yaml"
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/stringutil"
)

func combineVariables(a map[string]*yaml.Variable, b map[string]*yaml.Variable) map[string]*yaml.Variable {
	c := a

	for k, v := range b {
		c[k] = v
	}

	return c
}

func secretEnv(key string) string {
	return "$" + stringutil.Slugify(fmt.Sprintf("secret-%s", key))
}

// HandleSecrets handles the different 'Secret' arguments that are defined in the pipeline step.
// Secrets are given a generated value and placed in the 'environment', not a user-defined one. That value is then used when the pipeline attempts to retrieve the value in the argument.
// String arguments are already provided in the command line arguments when `cmdutil.StepCommand'
func HandleSecrets(c pipeline.Configurer, step pipeline.Step[pipeline.Action]) map[string]*yaml.Variable {
	env := map[string]*yaml.Variable{}
	for _, arg := range step.Arguments {
		switch arg.Type {
		case pipeline.ArgumentTypeSecret:
			env[secretEnv(arg.Key)] = &yaml.Variable{
				Secret: arg.Key,
			}
		}
	}

	return env
}

func stepVolumes(c pipeline.Configurer, step pipeline.Step[pipeline.Action]) []*yaml.VolumeMount {
	volumes := []*yaml.VolumeMount{}
	// TODO: It's unlikely that we want to actually associate volume mounts with "FS" type arguments.
	// We will probably want to zip those up and place them in the state volume or something...
	for _, v := range step.Arguments {
		if v.Type != pipeline.ArgumentTypeFS {
			continue
		}

		// Explicitely skip ArgumentSouceFS because it's available in every pipeline.
		if v == pipeline.ArgumentSourceFS {
			continue
		}

		// If it's a known argument...
		value, err := c.Value(v)
		if err != nil {
			// Skip this then because it's not known. It should be provided by a different step ran previously.
			// TODO: handle FS type arguments here?
		}

		volumes = append(volumes, &yaml.VolumeMount{
			Name:      stringutil.Slugify(v.Key),
			MountPath: value,
		})
	}

	return volumes
}

func NewStep(c pipeline.Configurer, path, state string, step pipeline.Step[pipeline.Action]) (*yaml.Container, error) {
	var (
		name    = stringutil.Slugify(step.Name)
		deps    = make([]string, len(step.Dependencies))
		image   = step.Image
		env     = map[string]*yaml.Variable{}
		volumes = stepVolumes(c, step)
	)

	for i, v := range step.Dependencies {
		deps[i] = stringutil.Slugify(v.Name)
	}

	cmd, err := cmdutil.StepCommand(c, cmdutil.CommandOpts{
		CompiledPipeline: PipelinePath,
		Path:             path,
		Step:             step,
		BuildID:          "$DRONE_BUILD_NUMBER",
		State:            state,
	})

	if err != nil {
		return nil, err
	}

	env = combineVariables(env, HandleSecrets(c, step))

	return &yaml.Container{
		Name:  name,
		Image: image,
		Commands: []string{
			strings.Join(cmd, " "),
		},
		DependsOn:   deps,
		Environment: env,
		Volumes:     volumes,
	}, nil
}
