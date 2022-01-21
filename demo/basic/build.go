package main

import (
	"log"

	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/plumbing/types"
)

func writeVersion(sw shipwright.Shipwright) types.StepAction {
	return func() error {
		log.Println("Writing version...")
		// equivalent of `git describe --tags --dirty --always`
		version := sw.Git.Describe(&git.DescribeOpts{
			Tags:   true,
			Dirty:  true,
			Always: true,
		})

		// write the version string in the `.version` file.
		return sw.FS.ReplaceString(".version", version)()
	}
}

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.New("basic pipeline", git.EventCommit{})
	defer sw.Done()

	// Define our steps so that we can refer to them in multiple places, like as the steps themselves or dependencies of other steps
	var (
		// Clone the project. Given that the git commit event is what triggered this pipeline,
		// the project details can be inferred.
		clone = sw.Git.Clone(1).WithName("clone")
	)

	sw.Run(clone)

	// In parallel, install the yarn and go dependencies, and cache the node_modules and $GOPATH/pkg folders.
	// The cache should invalidate if the yarn.lock or go.sum files have changed
	sw.Parallel(
		types.NamedStep("install frontend dependencies", sw.Cache(
			sw.Yarn.Install(),
			fs.Cache("node_modules", fs.FileHasChanged("yarn.lock")),
		)),
		types.NamedStep("install backend dependencies", sw.Cache(
			sw.Golang.Modules.Download(),
			fs.Cache("$GOPATH/pkg", fs.FileHasChanged("go.sum")),
		)),
	)

	sw.Run(
		types.NamedStep("write .version file", writeVersion(sw)),
		types.NamedStep("compile backend", sw.Make.Target("build")),
		types.NamedStep("compile frontend", sw.Make.Target("package")),
	)

	// sw.Output()
}
