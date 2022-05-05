package main

import (
	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/ci/docker"
	"github.com/grafana/shipwright/golang"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.NewMulti()
	defer sw.Done()

	sw.Run(
		sw.New("test and build", func(sw *shipwright.Shipwright[pipeline.Action]) {
			// Test the Golang code and ensure that the build steps
			sw.Run(
				golang.Test(sw, "./...").WithName("test"),
				docker.ShipwrightImage.BuildStep(sw).WithName("build shipwright docker image"),
			)

			sw.Run(docker.BuildSteps(sw, docker.Images)...)
		}),
	)

	sw.Run(
		sw.New("publish docker images", func(sw *shipwright.Shipwright[pipeline.Action]) {
			sw.When(
				pipeline.GitTagEvent(pipeline.GitTagFilters{}),
			)

			sw.Run(docker.Login(
				pipeline.NewSecretArgument("docker_username"),
				pipeline.NewSecretArgument("docker_password"),
				pipeline.NewSecretArgument("docker_registry"),
			))
			sw.Run(docker.BuildSteps(sw, docker.Images)...)
			sw.Run(docker.PushSteps(sw, docker.Images)...)
		}),
	)
}
