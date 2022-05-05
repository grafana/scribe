package shipwright

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/plumbing/stringutil"
)

func newFilesystemState(u *url.URL) (pipeline.State, error) {
	path := u.Path
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			path = filepath.Join(path, fmt.Sprintf("%s.json", stringutil.Random(8)))
		}
	}

	return pipeline.NewFilesystemState(path)
}

var states = map[string]func(u *url.URL) (pipeline.State, error){
	"file": newFilesystemState,
	"fs":   newFilesystemState,
}

func GetState(val string) (pipeline.State, error) {
	log.Println("got state", val)
	u, err := url.Parse(val)
	if err != nil {
		return nil, err
	}
	log.Println("got scheme", u.Scheme)

	if v, ok := states[u.Scheme]; ok {
		return v(u)
	}

	return nil, fmt.Errorf("state URL scheme '%s' not recognized", val)
}
