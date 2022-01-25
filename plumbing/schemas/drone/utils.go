package drone

import (
	"fmt"
	"strings"

	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/config"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

var (
	ErrorNoImage = plumbing.NewPipelineError("no image provided", "An image is required for all steps in Drone")
)

// Slugify removes illegal characters for use in identifiers in a Drone pipeline
func Slugify(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")

	return s
}

func Command(c config.Configurer, path string, step types.Step) (string, error) {
	args := make([]string, len(step.Arguments))

	for i, key := range step.Arguments {
		value, err := c.Value(key)
		if err != nil {
			return "", err
		}

		args[i] = fmt.Sprintf("-arg=%s=%s", string(key), value)
	}

	cmd := append([]string{"shipwright", fmt.Sprintf("-step=%d", step.Serial)}, args...)
	cmd = append(cmd, path)

	return strings.Join(cmd, " "), nil
}

func NewStep(c config.Configurer, path string, step types.Step) (Step, error) {
	if step.Image == "" {
		return Step{}, ErrorNoImage
	}

	var (
		name  = Slugify(step.Name)
		deps  = make([]string, len(step.Dependencies))
		image = step.Image
	)

	for i, v := range step.Dependencies {
		deps[i] = Slugify(v.Name)
	}

	cmd, err := Command(c, path, step)
	if err != nil {
		return Step{}, err
	}

	return Step{
		Name:  name,
		Image: image,
		Commands: []string{
			cmd,
		},
		DependsOn: deps,
	}, nil
}
