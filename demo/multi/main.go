package main

import "github.com/grafana/scribe"

var Pipelines = []scribe.Pipeline{
	PipelineDependencies,
	PipelineBuild,
	PipelineTest,
	PipelinePublish,
}

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the scribe client (sw).
// This allows the various clients to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := scribe.NewMulti()
	defer sw.Done()

	sw.AddPipelines(Pipelines...)
}
