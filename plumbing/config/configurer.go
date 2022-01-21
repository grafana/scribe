package config

import "pkg.grafana.com/shipwright/v1/plumbing/types"

type Configurer interface {
	// Value returns the implementation-specific pipeline config.
	Value(types.StepArgument) (string, error)
}
