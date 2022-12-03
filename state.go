package scribe

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/grafana/scribe/args"
	"github.com/grafana/scribe/state"
	"github.com/grafana/scribe/stringutil"
	"github.com/sirupsen/logrus"
)

func newFilesystemState(u *url.URL) (state.StateHandler, error) {
	path := u.Path
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			path = filepath.Join(path, fmt.Sprintf("%s.json", stringutil.Random(8)))
		}
	}

	return state.NewFilesystemState(path)
}

var states = map[string]func(*url.URL) (state.StateHandler, error){
	"file": newFilesystemState,
	"fs":   newFilesystemState,
}

func GetState(val string, log logrus.FieldLogger, pargs *args.PipelineArgs) (*state.State, error) {
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}

	fallback := []state.StateReader{
		state.StateReaderWithLogs(log.WithField("state", "arguments"), state.NewArgMapReader(pargs.ArgMap)),
	}

	if pargs.CanStdinPrompt {
		fallback = append(fallback, state.StateReaderWithLogs(log.WithField("state", "stdin"), state.NewStdinReader(os.Stdin, os.Stdout)))
	}

	if v, ok := states[u.Scheme]; ok {
		handler, err := v(u)
		if err != nil {
			return nil, err
		}

		return &state.State{
			Handler:  state.StateHandlerWithLogs(log.WithField("state", u.Scheme), handler),
			Fallback: fallback,
			Log:      log,
		}, nil
	}

	return nil, fmt.Errorf("state URL scheme '%s' not recognized", val)
}
