package drone

import (
	"fmt"

	"github.com/drone/drone-yaml/yaml"
	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/cmdutil"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/scribe/plumbing/stringutil"
	"github.com/sirupsen/logrus"
)

func combineVariables(a map[string]*yaml.Variable, b map[string]*yaml.Variable) map[string]*yaml.Variable {
	c := a

	for k, v := range b {
		c[k] = v
	}

	return c
}

func secretEnv(key string) string {
	return stringutil.Slugify(fmt.Sprintf("secret_%s", key))
}

// HandleSecrets handles the different 'Secret' arguments that are defined in the pipeline step.
// Secrets are given a generated value and placed in the 'environment', not a user-defined one. That value is then used when the pipeline attempts to retrieve the value in the argument.
// String arguments are already provided in the command line arguments when `cmdutil.StepCommand'
func HandleSecrets(c pipeline.Configurer, step pipeline.Step) (map[string]*yaml.Variable, map[string]string) {
	var (
		env  = make(map[string]*yaml.Variable)
		args = make(map[string]string)
	)

	for _, arg := range step.Arguments {
		name := secretEnv(arg.Key)
		switch arg.Type {
		case pipeline.ArgumentTypeSecret:
			env[name] = &yaml.Variable{
				Secret: arg.Key,
			}
			args[arg.Key] = "$" + name
		}
	}

	return env, args
}

func stepVolumes(c pipeline.Configurer, step pipeline.Step) []*yaml.VolumeMount {
	volumes := []*yaml.VolumeMount{}
	// TODO: It's unlikely that we want to actually associate volume mounts with "FS" type arguments.
	// We will probably want to zip those up and place them in the state volume or something...
	for _, v := range step.Arguments {
		if v.Type != pipeline.ArgumentTypeFS && v.Type != pipeline.ArgumentTypeUnpackagedFS {
			continue
		}

		// Explicitely skip ArgumentSouceFS because it's available in every pipeline.
		if v == pipeline.ArgumentSourceFS {
			continue
		}

		// If it's a known argument...
		value, _ := c.Value(v)
		//if err != nil {
		// Skip this then because it's not known. It should be provided by a different step ran previously.
		// TODO: handle FS type arguments here?
		//}

		volumes = append(volumes, &yaml.VolumeMount{
			Name:      stringutil.Slugify(v.Key),
			MountPath: value,
		})
	}

	return volumes
}

func NewDaggerStep(c pipeline.Configurer, path, state, version string, p pipeline.Pipeline) (*yaml.Container, error) {
	var (
		name  = stringutil.Slugify(p.Name)
		image = "go:1.19"
		//volumes = stepVolumes(c, step)
	)
	//env, args := HandleSecrets(c, p)

	//for i, v := range step.Dependencies {
	//	deps[i] = stringutil.Slugify(v.Name)
	//}

	cmd, err := cmdutil.PipelineCommand(cmdutil.PipelineCommandOpts{
		Pipeline: p,
		CommandOpts: cmdutil.CommandOpts{
			CompiledPipeline: PipelinePath,
			PipelineArgs: plumbing.PipelineArgs{
				Path:    path,
				BuildID: "$DRONE_BUILD_NUMBER",
				State:   state,
				//ArgMap:        args,
				LogLevel: logrus.DebugLevel,
				Version:  version,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return &yaml.Container{
		Name:     name,
		Image:    image,
		Commands: cmd,
	}, nil
}
