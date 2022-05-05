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
	return stringutil.Slugify(fmt.Sprintf("secret-%s", key))
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

func NewStep(c pipeline.Configurer, path, state string, step pipeline.Step[pipeline.Action]) (*yaml.Container, error) {
	var (
		name  = stringutil.Slugify(step.Name)
		deps  = make([]string, len(step.Dependencies))
		image = step.Image
		env   = map[string]*yaml.Variable{}
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
	}, nil
}
