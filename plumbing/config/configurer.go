package config

import "pkg.grafana.com/shipwright/v1/plumbing/pipeline"

type Configurer interface {
	// Value returns the implementation-specific pipeline config.
	Value(pipeline.StepArgument) (string, error)
}
