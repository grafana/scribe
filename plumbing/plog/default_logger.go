package plog

import (
	"os"
)

// DefaultLogger is used in the package-level log functions. It uses the flag package to check for a -log-level flag.
var DefaultLogger = New(LogLevelDebug, os.Stderr)
