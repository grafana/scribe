package main

import (
	"context"
	"regexp"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/fs"
	gitx "github.com/grafana/shipwright/git/x"
	"github.com/grafana/shipwright/golang"
	"github.com/grafana/shipwright/makefile"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/yarn"
)

func writeVersion(sw shipwright.Shipwright) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {

		// equivalent of `git describe --tags --dirty --always`
		version, err := gitx.Describe(ctx, ".", true, true, true)
		if err != nil {
			return err
		}

		// write the version string in the `.version` file.
		return fs.ReplaceString(".version", version)(ctx, opts)
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

	sw.When(
		pipeline.GitCommitEvent(pipeline.GitCommitFilters[string]{
			Branch: pipeline.StringFilter("main"),
		}),
		pipeline.GitTagEvent(pipeline.GitTagFilters[*regexp.Regexp]{
			Name: pipeline.RegexpFilter(regexp.MustCompile("^v([0-9]).*$")),
		}),
	)

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
