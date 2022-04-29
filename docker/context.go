package docker

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/archive"
)

// ImageContext creates a tar directory for the given directory (dir) that respects the '.dockerignore' file located at '$dir/.dockerignore'.
func ImageContext(dir string) (io.ReadCloser, error) {
	ignore := []string{}
	f := filepath.Join(dir, ".dockerignore")
	if _, err := os.Stat(f); err == nil {
		ignoref, err := os.Open(f)
		if err != nil {
			return nil, fmt.Errorf("failed to open dockerignore '%s': %w", f, err)
		}

		i, err := dockerignore.ReadAll(ignoref)
		if err != nil {
			return nil, fmt.Errorf("failed to read dockerignore '%s': %w", f, err)
		}

		ignore = i
	}

	return archive.TarWithOptions(dir, &archive.TarOptions{
		ExcludePatterns: ignore,
	})
}
