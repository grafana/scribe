package shipwright

import "flag"

type RunMode int

const (
	RunModeCLI RunMode = iota
	RunModeServer
	RunModeConfig
)

func (r *RunMode) String() string {
	switch *r {
	case RunModeCLI:
		return "cli"
	case RunModeServer:
		return "server"
	case RunModeConfig:
		return "config"
	}

	return "unknown"
}

func (r *RunMode) Set(v string) error {
	switch v {
	case "server":
		*r = RunModeServer
	case "config":
		*r = RunModeConfig
	default:
		*r = RunModeCLI
	}

	return nil
}

type Opts struct {
	Mode RunMode
}

func ParseCLIOpts(args []string) (*Opts, error) {
	opts := &Opts{}

	flagSet := flag.NewFlagSet("shipwright", flag.ExitOnError)
	flagSet.Var(&opts.Mode, "mode", "An internal flag that defines what mode the shipwright pipeline is running in.")

	if err := flagSet.Parse(args); err != nil {
		return nil, err
	}
	return opts, nil
}
