package shipwright

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
)

type RunMode struct {
	Opt    RunModeOption
	Client Shipwright
}

// String outputs the string equivelant of what mode is selected.
func (r *RunMode) String() string {
	switch r.Opt {
	case RunModeCLI:
		return "cli"
	case RunModeServer:
		return "server"
	case RunModeConfig:
		return "config"
	case RunModeDrone:
		return "drone"
	}

	return "unknown"
}

// Set sets the active run mode and the handler based on "v"
func (r *RunMode) Set(v string) error {
	switch v {
	case "server":
		r.Opt = RunModeServer
	case "config":
		r.Opt = RunModeConfig
	default:
		r.Opt = RunModeCLI
		r.Client = NewCLIClient()
	}

	return nil
}
