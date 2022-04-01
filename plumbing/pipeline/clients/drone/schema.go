package drone

import (
	"strings"

	"github.com/drone/drone-yaml/yaml"
	"github.com/grafana/shipwright/plumbing/cmdutil"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/stringutil"
)

func NewStep(c pipeline.Configurer, path string, step pipeline.Step) (*yaml.Container, error) {
	var (
		name  = stringutil.Slugify(step.Name)
		deps  = make([]string, len(step.Dependencies))
		image = step.Image
	)

	for i, v := range step.Dependencies {
		deps[i] = stringutil.Slugify(v.Name)
	}

	cmd, err := cmdutil.StepCommand(c, cmdutil.CommandOpts{
		Path:    path,
		Step:    step,
		BuildID: "$DRONE_BUILD_NUMBER",
	})

	if err != nil {
		return nil, err
	}

	return &yaml.Container{
		Name:  name,
		Image: image,
		Commands: []string{
			strings.Join(cmd, " "),
		},
		DependsOn: deps,
	}, nil
}
