package main

import (
	"time"

	"github.com/grafana/shipwright"
	"github.com/grafana/shipwright/plumbing/pipeline"
)

func main() {
	sw := shipwright.New("complex-pipeline")
	defer sw.Done()

	sw.Run(pipeline.NamedStep("initalize", NoOpAction("initialize", time.Second*22)))

	sw.Parallel(
		pipeline.NamedStep("build backend", NoOpAction("buildbackend", time.Second*39)),
		pipeline.NamedStep("build frontend", NoOpAction("buildfrontend", time.Minute)),
		pipeline.NamedStep("build documentation", NoOpAction("builddocs", time.Second*9)),
	)

	sw.Parallel(
		pipeline.NamedStep("test backend", NoOpAction("testbackend", time.Second*27)),
		pipeline.NamedStep("test frontend", NoOpAction("testfrontend", time.Second*32)),
	)

	sw.Run(
		pipeline.NamedStep("integration tests: sqlite", NoOpAction("integrationtests_sqlite", time.Minute)),
		pipeline.NamedStep("integration tests: postgres", NoOpAction("integrationtests_pg", time.Second*42)),
		pipeline.NamedStep("integration tests: mysql", NoOpAction("integrationtests_mysql", time.Second*32)),
		pipeline.NamedStep("integration tests: mssql", NoOpAction("integrationtests_mssql", time.Second*55)),
	)

	sw.Run(
		pipeline.NamedStep("package", NoOpAction("package", time.Second*13)),
		pipeline.NamedStep("build docker image", NoOpAction("builddocker", time.Second*44)),
	)

	sw.Parallel(
		pipeline.NamedStep("publish documentation", NoOpAction("publishdocs", time.Second*13)),
		pipeline.NamedStep("publish package", NoOpAction("publishpackage", time.Second*12)),
		pipeline.NamedStep("publish docker image", NoOpAction("publishdockerimage", time.Second*23)),
	)

	sw.Parallel(
		pipeline.NamedStep("notify slack", NoOpAction("notifyslack", time.Second*3)),
	)
}