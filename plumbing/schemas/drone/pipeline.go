package drone

type EnvironmentVariable struct{}

type CloneSettings struct {
	Disable bool `yaml:"disable"`
}

// Step is a single step in Drone in a docker pipeline
type Step struct {
	Name        string                         `yaml:"name,omitempty"`
	Image       string                         `yaml:"image,omitempty"`
	Commands    []string                       `yaml:"commands,omitempty"`
	DependsOn   []string                       `yaml:"depends_on,omitempty"`
	Environment map[string]EnvironmentVariable `yaml:"environment,omitempty"`
}

// Pipeline represents a single pipeline in the `.drone.yml` configuration file.
type Pipeline struct {
	Kind        string                         `yaml:"kind,omitempty"`
	Type        string                         `yaml:"type,omitempty"`
	Name        string                         `yaml:"name,omitempty"`
	Clone       CloneSettings                  `yaml:"clone"`
	Environment map[string]EnvironmentVariable `yaml:"environment,omitempty"`
	Steps       []Step                         `yaml:"steps,omitempty"`
}

type Config []Pipeline
