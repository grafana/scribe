package plumbing

import "fmt"

const DefaultRepo = "grafana/shipwright"

func DefaultImage(version string) string {
	// TODO don't hardcode this image but for now I don't care good luck
	return fmt.Sprintf("%s:%s", DefaultRepo, version)
}

func SubImage(image, version string) string {
	return fmt.Sprintf("%s:%s-%s", DefaultRepo, version, image)
}
