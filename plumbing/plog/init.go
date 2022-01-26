package plog

import (
	"flag"
	"io"
	"os"
)

// DefaultLogger is used in the package-level log functions. It uses the flag package to check for a -log-level flag.
var DefaultLogger = New(LogLevelInfo, os.Stderr)

func init() {
	f := flag.NewFlagSet("shipwright logging", flag.ContinueOnError)

	level := LogLevel(1)

	f.Var(&level, "log-level", "debug|info|warn|error")

	f.SetOutput(io.Discard)
	f.Parse(os.Args)

	DefaultLogger.level = level
}
