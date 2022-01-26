package main

import (
	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/ci/docker"
	"pkg.grafana.com/shipwright/v1/git"
)

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.New("basic pipeline", git.EventCommit{})
	defer sw.Done()

	sw.Run(
		sw.Git.Clone(1).WithName("clone"),
		sw.Golang.Test().WithName("test"),
		docker.ShipwrightImage.BuildStep(sw).WithName("build shipwright docker image"),
	)

	// Build all of the shipwright docker images in parallel
	// With unbound parallelism this could cause some very poor performance
	sw.Parallel(docker.Steps(sw, docker.Images)...)
}
