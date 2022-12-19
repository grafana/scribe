package state

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/grafana/scribe/args"
	"github.com/sirupsen/logrus"
)

// newFilesystemState creates a new filesystem state handler.
// If the directory provided doesn't exist, it will be created.
func newFilesystemState(ctx context.Context, u *url.URL) (Handler, error) {
	var (
		dir = u.Path
	)

	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("state argument '%s' must be a directory. example: '/var/scribe-state'", dir)
		}
	} else {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("state directory '%s' does not exist. Error attempting to create it: %w", dir, err)
			}
		}
	}

	return NewFilesystemState(dir)
}

func newGCSState(ctx context.Context, u *url.URL) (Handler, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return NewGCSHandler(client, u)
}

func newS3State(ctx context.Context, u *url.URL) (Handler, error) {
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(sdkConfig)
	return NewS3Handler(client, u)

}

var states = map[string]func(context.Context, *url.URL) (Handler, error){
	"file": newFilesystemState,
	"fs":   newFilesystemState,
	"gs":   newGCSState,
	"gcs":  newGCSState,
	"s3":   newS3State,
}

// NewDefaultState creates a new default state given the arguments provided.
// The --no-stdin flag will prevent the State object from using the stdin to populate the state for ClientProvidedArguments. (See `pipeline/arguments_known.go` for those).
// The --state flag defines where the state JSON and state data will be stored.
// If the value for a key is not available in the primary state (defined by the --state flag), then the state object will attempt to retrieve it from the fallback. Currently, the fallback options are the `--arg` flags (--arg={key}={value}), or, if `--no-stdin` is not set, then from the stdin.
func NewDefaultState(ctx context.Context, log logrus.FieldLogger, pargs *args.PipelineArgs) (*State, error) {
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
		handler, err := v(ctx, u)
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
