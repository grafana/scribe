package main

import (
	"github.com/grafana/scribe"
	"github.com/grafana/scribe/ci/docker"
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
				golang.Test(sw, "./...").WithName("test"),
				docker.ScribeImage.BuildStep(sw).WithName("build scribe docker image"),
			)

			sw.Run(docker.BuildSteps(sw, docker.Images)...)
		}),
	)

	sw.Run(
		sw.New("publish docker images", func(sw *scribe.Scribe) {
			// sw.When(
			// 	pipeline.GitTagEvent(pipeline.GitTagFilters{}),
			// )

			login := docker.Login(
				pipeline.NewSecretArgument("docker_username"),
				pipeline.NewSecretArgument("docker_password"),
			).
				WithName("docker login").
				WithImage(plumbing.SubImage("docker", sw.Version))

			sw.Run(login)

			sw.Run(docker.BuildSteps(sw, docker.Images)...)
			sw.Run(docker.ScribeImage.PushStep(sw).WithName("push scribe docker image"))
			sw.Run(docker.PushSteps(sw, docker.Images)...)
		}),
	)
}
