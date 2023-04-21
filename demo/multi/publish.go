package main

import (
	"context"
	"os"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

var (
	ArgumentTarPackage = state.NewFileArgument("package-tarball")
)

func actionPackage(ctx context.Context, opts pipeline.ActionOpts) error {
	f, err := os.CreateTemp("", "*.tar.gz")
	if err != nil {
		return err
	}

	opts.Logger.Infoln("Created file:", f.Name())
	defer f.Close()
	path, err := opts.State.SetFileReader(ctx, ArgumentTarPackage, f)
	opts.Logger.Infoln("Stored in state as:", path)
	return err
}

func actionPublish(ctx context.Context, opts pipeline.ActionOpts) error {
	opts.Logger.Warnln("Pipeline done!")
	opts.Logger.Warnln("Pipeline done!")
	opts.Logger.Warnln("Pipeline done!")

	return nil
}

var stepPackage = pipeline.NamedStep("package", actionPackage).
	Provides(ArgumentTarPackage).
	Requires(ArgumentCompiledBackend, ArgumentCompiledFrontend)

var stepPublish = pipeline.NamedStep("publish", actionPublish).
	Requires(
		state.NewSecretArgument("gcp-publish-key"),
		ArgumentTarPackage,
	)

var PipelinePublish = scribe.Pipeline{
	Name: "publish",
	Steps: []pipeline.Step{
		stepPackage,
		stepPublish,
	},
	Requires: []state.Argument{ArgumentCompiledBackend, ArgumentCompiledFrontend},
	Provides: []state.Argument{ArgumentTarPackage},
	When: []pipeline.Event{
		pipeline.GitCommitEvent(pipeline.GitCommitFilters{
			Branch: pipeline.StringFilter("main"),
		}),
		pipeline.GitTagEvent(pipeline.GitTagFilters{
			Name: pipeline.GlobFilter("v*"),
		}),
	},
}
