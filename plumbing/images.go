package plumbing

import "fmt"

func DefaultImage(version string) string {
	// TODO don't hardcode this image but for now I don't care good luck
	return fmt.Sprintf("ghcr.io/grafana/shipwright:%s", version)
}

func SubImage(image, version string) string {
	return fmt.Sprintf("ghcr.io/grafana/shipwright/%s:%s", image, version)
}
