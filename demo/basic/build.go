package main

import (
	"context"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/fs"
	gitx "github.com/grafana/scribe/git/x"
	"github.com/grafana/scribe/golang"
	"github.com/grafana/scribe/makefile"
	"github.com/grafana/scribe/plumbing/pipeline"
	"github.com/grafana/scribe/yarn"
)

func writeVersion(sw *scribe.Scribe) pipeline.Step {
	action := func(ctx context.Context, opts pipeline.ActionOpts) error {

		// equivalent of `git describe --tags --dirty --always`
		version, err := gitx.Describe(ctx, ".", true, true, true)
		if err != nil {
			return err
		}

		// write the version string in the `.version` file.
		return fs.ReplaceString(".version", version)(ctx, opts)
	}

	return pipeline.NewStep(action).WithImage("alpine:latest")
}

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the scribe client (sw).
// This allows the various clients to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := scribe.New("basic pipeline")
	defer sw.Done()

	sw.When(
		pipeline.GitCommitEvent(pipeline.GitCommitFilters{
			Branch: pipeline.StringFilter("main"),
		}),
		pipeline.GitTagEvent(pipeline.GitTagFilters{
			Name: pipeline.GlobFilter("v*"),
		}),
	)

	// In parallel, install the yarn and go dependencies, and cache the node_modules and $GOPATH/pkg folders.
	// The cache should invalidate if the yarn.lock or go.sum files have changed
	sw.Run(
		pipeline.NamedStep("install frontend dependencies", sw.Cache(
			yarn.InstallAction(),
			fs.Cache("node_modules", fs.FileHasChanged("yarn.lock")),
		)).WithImage("node:latest"),
		pipeline.NamedStep("install backend dependencies", sw.Cache(
			golang.ModDownload(),
			fs.Cache("$GOPATH/pkg", fs.FileHasChanged("go.sum")),
		)).WithImage("node:latest"),
		writeVersion(sw).WithName("write-version-file"),
	)

	sw.Run(
		pipeline.NamedStep("compile backend", makefile.Target("build")).WithImage("alpine:latest"),
		pipeline.NamedStep("compile frontend", makefile.Target("package")).WithImage("alpine:latest"),
		pipeline.NamedStep("build docker image", makefile.Target("build")).WithArguments(pipeline.ArgumentDockerSocketFS).WithImage("alpine:latest"),
	)

	sw.Run(
		pipeline.NamedStep("publish", makefile.Target("publish")).
			WithArguments(
				pipeline.NewSecretArgument("gcs-publish-key"),
			),
	)
}
