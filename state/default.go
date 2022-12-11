package state

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/stringutil"
	"github.com/sirupsen/logrus"
)

func newFilesystemState(u *url.URL) (Handler, error) {
	path := u.Path
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			path = filepath.Join(path, fmt.Sprintf("%s.json", stringutil.Random(8)))
		}
	}

	return NewFilesystemState(path)
}

var states = map[string]func(*url.URL) (Handler, error){
	"file": newFilesystemState,
	"fs":   newFilesystemState,
}

// NewDefaultState creates a new default state given the arguments provided.
// The --no-stdin flag will prevent the State object from using the stdin to populate the state for ClientProvidedArguments. (See `pipeline/arguments_known.go` for those).
// The --state flag defines where the state JSON and state data will be stored.
// If the value for a key is not available in the primary state (defined by the --state flag), then the state object will attempt to retrieve it from the fallback. Currently, the fallback options are the `--arg` flags (--arg={key}={value}), or, if `--no-stdin` is not set, then from the stdin.
func NewDefaultState(log logrus.FieldLogger, pargs *args.PipelineArgs) (*State, error) {
	u, err := url.Parse(pargs.State)
	if err != nil {
		return nil, err
	}

	fallback := []Reader{
		ReaderWithLogs(log.WithField("state", "arguments"), NewArgMapReader(pargs.ArgMap)),
	}

	if pargs.CanStdinPrompt {
		fallback = append(fallback, ReaderWithLogs(log.WithField("state", "stdin"), NewStdinReader(os.Stdin, os.Stdout)))
	}

	if v, ok := states[u.Scheme]; ok {
		handler, err := v(u)
		if err != nil {
			return nil, err
		}

		return &State{
			Handler:  HandlerWithLogs(log.WithField("state", u.Scheme), handler),
			Fallback: fallback,
			Log:      log,
		}, nil
	}

	return nil, fmt.Errorf("state URL scheme '%s' not recognized", pargs.State)
}
