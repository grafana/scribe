package drone

import (
	"fmt"
	"strings"

	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

// Slugify removes illegal characters for use in identifiers in a Drone pipeline
func Slugify(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")

	return s
}

func NewStep(path string, step types.Step) Step {
	name := Slugify(step.Name)
	deps := make([]string, len(step.Dependencies))
	image := "grafana/shipwright:latest"

	for i, v := range step.Dependencies {
		deps[i] = Slugify(v.Name)
	}

	if step.Image != "" {
		image = step.Image
	}

	return Step{
		Name:  name,
		Image: image,
		Commands: []string{
			fmt.Sprintf("shipwright -step=%d %s", step.Serial, path),
		},
		DependsOn: deps,
	}
}
