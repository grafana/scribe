package main

import (
	"pkg.grafana.com/shipwright/v1"
	"pkg.grafana.com/shipwright/v1/fs"
	gitx "pkg.grafana.com/shipwright/v1/git/x"
	"pkg.grafana.com/shipwright/v1/golang"
	"pkg.grafana.com/shipwright/v1/makefile"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/yarn"
)

func writeVersion(sw shipwright.Shipwright) pipeline.Step {
	action := func(opts pipeline.ActionOpts) error {

		// equivalent of `git describe --tags --dirty --always`
		version, err := gitx.Describe(".", true, true, true)
		if err != nil {
			return err
		}

		// write the version string in the `.version` file.
		return fs.ReplaceString(".version", version)(opts)
	}

	return pipeline.NewStep(action)
}

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.New("basic pipeline")
	defer sw.Done()

	// In parallel, install the yarn and go dependencies, and cache the node_modules and $GOPATH/pkg folders.
	// The cache should invalidate if the yarn.lock or go.sum files have changed
	sw.Run(
		pipeline.NamedStep("install frontend dependencies", sw.Cache(
			yarn.Install(),
			fs.Cache("node_modules", fs.FileHasChanged("yarn.lock")),
		)),
		pipeline.NamedStep("install backend dependencies", sw.Cache(
			golang.ModDownload(),
			fs.Cache("$GOPATH/pkg", fs.FileHasChanged("go.sum")),
		)),
		writeVersion(sw).WithName("write-version-file"),
	)

	sw.Run(
		pipeline.NamedStep("compile backend", makefile.Target("build")),
		pipeline.NamedStep("compile frontend", makefile.Target("package")),
	)
}
