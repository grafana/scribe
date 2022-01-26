package docker

import (
	"embed"
)

//go:embed *.Dockerfile
var Dockerfiles embed.FS
