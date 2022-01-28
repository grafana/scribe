package plumbing

import (
	"errors"
)

// RunModeOption defines a secenario in which Shipwright can process a pipeline
type RunModeOption int

const (
	// RunModeCLI is set when a pipeline is ran from the Shipwright CLI, typically for local development, but can also be set when running Shipwright within a third-party service like CircleCI or Drone
	RunModeCLI RunModeOption = iota

	// RunModeServer is set when a pipeline is ran from the Shipwright server.
	RunModeServer

	// RunModeConfig is set when a pipeline is ran in configuration / metadata mode. Typically this is used in tandem with the Server mode.
	RunModeConfig

	// RunModeDrone is set when a pipeline is ran in Drone mode, which is used to generate a Drone config from a Shipwright pipeline
	RunModeDrone

	// RunModeDocker runs the pipeline using the Docker CLI for each step
	RunModeDocker
)

var runModeStr = []string{"cli", "server", "config", "drone", "docker"}

// String outputs the string equivelant of what mode is selected.
func (r *RunModeOption) String() string {
	if int(*r) >= len(runModeStr) {
		return "unknown"
	}

	return runModeStr[*r]
}

// Set sets the active run mode and the handler based on "v"
func (r *RunModeOption) Set(val string) error {
	for i, v := range runModeStr {
		if v == val {
			*r = RunModeOption(i)
			return nil
		}
	}

	return errors.New("unknown mode")
}
