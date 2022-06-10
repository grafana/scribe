package main

import (
	"github.com/grafana/scribe"
	"github.com/grafana/scribe/golang"
	"github.com/grafana/scribe/plumbing"
	"github.com/grafana/scribe/plumbing/pipeline"
)

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the scribe client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := scribe.NewMulti()
	defer sw.Done()

	sw.Run(
		sw.New("test and build", func(sw *scribe.Scribe) {
			// Test the Golang code and ensure that the build steps
			sw.Run(
				StepGetVersion(sw.Version).WithName("get version"),
				golang.Test(sw, "./...").WithName("test"),
				StepBuildImage(sw.Version, ScribeImage).WithName("build scribe docker image"),
			)

			sw.Run(BuildSteps(sw.Version, Images)...)
		}),
	)

	sw.Run(
		sw.New("publish docker images", func(sw *scribe.Scribe) {
			// sw.When(
			// 	pipeline.GitTagEvent(pipeline.GitTagFilters{}),
			// )

			login := StepDockerLogin(
				pipeline.NewSecretArgument("docker_username"),
				pipeline.NewSecretArgument("docker_password"),
			).
				WithName("docker login").
				WithImage(plumbing.SubImage("docker", sw.Version))

			sw.Run(
				login,
				StepGetVersion(sw.Version).WithName("get version"),
			)

			sw.Run(BuildSteps(sw.Version, Images)...)
			sw.Run(ListImages().WithName("list images"))
			sw.Run(StepPushImage(sw.Version, ScribeImage).WithName("push scribe docker image"))
			sw.Run(PushSteps(sw.Version, Images)...)
		}),
	)
}
