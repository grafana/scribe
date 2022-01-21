package main

import (
	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
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
		types.NamedStep("test", sw.Golang.Test()),
	)
}
