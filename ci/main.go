package main

import (
	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/ci/docker"
	"github.com/grafana/shipwright/golang"
)

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.New("groan")
	defer sw.Done()

	sw.Run(
		golang.Test(sw, "./...").WithName("test"),
		docker.ShipwrightImage.BuildStep(sw).WithName("build shipwright docker image"),
	)

	// Build all of the shipwright docker images in parallel
	// With unbound parallelism this could cause some very poor performance
	sw.Run(docker.BuildSteps(sw, docker.Images)...)

	// sw.Run(docker.PushSteps(sw, docker.Images)...)
}
