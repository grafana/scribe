package drone

import (
	"strings"

	"pkg.grafana.com/shipwright/v1/plumbing/cmdutil"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/stringutil"
)

func NewStep(c pipeline.Configurer, path string, step pipeline.Step) (Step, error) {
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
		return Step{}, err
	}

	return Step{
		Name:  name,
		Image: image,
		Commands: []string{
			strings.Join(cmd, " "),
		},
		DependsOn: deps,
	}, nil
}
