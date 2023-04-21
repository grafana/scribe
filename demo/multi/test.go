package main

import (
	"context"
	"time"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
	"github.com/grafana/scribe/state"
)

var (
	ArgumentTestResultBackend  = state.NewBoolArgument("backend-test-result")
	ArgumentTestResultFrontend = state.NewBoolArgument("frontend-test-result")
)

func actionTestFrontend(ctx context.Context, opts pipeline.ActionOpts) error {
	opts.Logger.Infoln("Testing frontend...")
	time.Sleep(time.Second * 1)
	opts.Logger.Infoln("Done testing frontend")
	// make test-frontend
	// assume it passed...
	return opts.State.SetBool(ctx, ArgumentTestResultFrontend, true)
}

func actionTestBackend(ctx context.Context, opts pipeline.ActionOpts) error {
	opts.Logger.Infoln("Testing backend...")
	time.Sleep(time.Second * 1)
	opts.Logger.Infoln("Done testing backend")
	// go test ./...
	// assume it passed...
	return opts.State.SetBool(ctx, ArgumentTestResultBackend, true)
}

var stepTestBackend = pipeline.NamedStep("test backend", actionTestBackend).
	Provides(ArgumentTestResultBackend).
	Requires(ArgumentGoDependencies)

var stepTestFrontend = pipeline.NamedStep("test frontend", actionTestFrontend).
	Provides(ArgumentTestResultFrontend).
	Requires(ArgumentNodeDependencies)

var PipelineTest = scribe.Pipeline{
	Name: "test",
	Steps: []pipeline.Step{
		stepTestBackend,
		stepTestFrontend,
	},
	Requires: []state.Argument{ArgumentNodeDependencies, ArgumentGoDependencies},
	Provides: []state.Argument{ArgumentTestResultFrontend, ArgumentTestResultBackend},
}
