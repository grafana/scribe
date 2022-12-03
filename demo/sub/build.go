package main

import (
	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
)

// "main" defines our program pipeline.
// Every pipeline step should be instantiated using the scribe client (sw).
// This allows the various clients to work properly in different scenarios, like in a CI environment or locally.
// Logic and processing done outside of the `sw.*` family of functions may not be included in the resulting pipeline.
func main() {
	sw := scribe.New("demo-pipeline-with-sub")
	defer sw.Done()

	sw.Sub(func(sw *scribe.Scribe) {
		sw.Run(pipeline.NoOpStep.WithName("sub-step-1"))
		sw.Parallel(
			pipeline.NoOpStep.WithName("sub-step-2"),
			pipeline.NoOpStep.WithName("sub-step-3"),
		)
	})

	sw.Run(
		pipeline.NoOpStep.WithName("step-1"),
		pipeline.NoOpStep.WithName("step-2"),
		pipeline.NoOpStep.WithName("step-3"),
	)

	sw.Parallel(
		pipeline.NoOpStep.WithName("step-4"),
		pipeline.NoOpStep.WithName("step-5"),
		pipeline.NoOpStep.WithName("step-6"),
	)
}
