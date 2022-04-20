package main

import (
	"context"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/fs"
	gitx "github.com/grafana/shipwright/git/x"
	"github.com/grafana/shipwright/golang"
	"github.com/grafana/shipwright/makefile"
	"github.com/grafana/shipwright/plumbing/pipeline"
	"github.com/grafana/shipwright/yarn"
)

func writeVersion(sw *shipwright.Shipwright[pipeline.Action]) pipeline.Step[pipeline.Action] {
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

func installDependencies(sw *shipwright.Shipwright[pipeline.Action]) {
	sw.Run(
		pipeline.NamedStep("install frontend dependencies", sw.Cache(
			yarn.Install(),
			fs.Cache("node_modules", fs.FileHasChanged("yarn.lock")),
		)),
		pipeline.NamedStep("install backend dependencies", sw.Cache(
			golang.ModDownload(),
			fs.Cache("$GOPATH/pkg", fs.FileHasChanged("go.sum")),
		)),
	)
}

func testPipeline(sw *shipwright.Shipwright[pipeline.Action]) {
	installDependencies(sw)

	sw.Parallel(
		golang.Test(sw, "./...").WithName("test backend"),
		pipeline.NamedStep("test frontend", makefile.Target("test-frontend")),
	)
}

func publishPipeline(sw *shipwright.Shipwright[pipeline.Action]) {
	sw.When(
		pipeline.GitCommitEvent(pipeline.GitCommitFilters{
			Branch: pipeline.StringFilter("main"),
		}),
		pipeline.GitTagEvent(pipeline.GitTagFilters{
			Name: pipeline.GlobFilter("v*"),
		}),
	)

	installDependencies(sw)

	sw.Parallel(
		pipeline.NamedStep("compile backend", makefile.Target("build")),
		pipeline.NamedStep("compile frontend", makefile.Target("package")),
	)

	sw.Run(
		pipeline.NamedStep("publish", makefile.Target("publish")).WithArguments(pipeline.NewSecretArgument("gcp-publish-key")),
	)
}

func codeqlPipeline(sw *shipwright.Shipwright[pipeline.Action]) {
	sw.Run(
		pipeline.NoOpStep.WithName("codeql"),
		pipeline.NoOpStep.WithName("notify-slack"),
	)
}

// "main" defines our program pipeline.
// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the shipwright client (sw).
// This allows the various client modes to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := shipwright.NewMulti()
	defer sw.Done()

	// Presumably this function could run for 10+ minutes so we want to run it while test & publish are happening.
	// We could run it in parallel with test, and if it fails, don't run publish, but we don't actually care if this passes in order to publish a pre-release build.
	sw.Sub(func(sw *shipwright.Shipwright[pipeline.Pipeline]) {
		sw.Run(sw.New("code quality check", codeqlPipeline))
	})

	sw.Run(
		sw.New("test", testPipeline),
		sw.New("publish", publishPipeline),
	)
}
