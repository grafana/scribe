package main

import (
	"context"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/golang"
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
			sw.Run(golang.Test(sw, "./...").WithName("test"))
		}),
	)

	sw.Run(
		sw.New("create github release", func(sw *scribe.Scribe) {
			sw.When(
				pipeline.GitTagEvent(pipeline.GitTagFilters{}),
			)

			sw.Run(pipeline.NamedStep("am I on a tag event?", func(ctx context.Context, opts pipeline.ActionOpts) error {
				opts.Logger.Infoln("1. I'm on a tag event.")
				opts.Logger.Infoln("2. I'm on a tag event.")
				opts.Logger.Infoln("3. I'm on a tag event.")
				return nil
			}).WithImage("alpine:latest"))
		}),
	)
}
