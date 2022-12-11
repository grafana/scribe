package main

import (
	"time"

	"github.com/grafana/scribe"
	"github.com/grafana/scribe/pipeline"
)

func main() {
	sw := scribe.New("complex-pipeline")
	defer sw.Done()

	sw.Background(pipeline.NamedStep("redis", pipeline.DefaultAction).WithImage("redis:6"))

	sw.Add(pipeline.NamedStep("initalize", NoOpAction("initialize", time.Second*22)))

	sw.Add(
		pipeline.NamedStep("build backend", NoOpAction("buildbackend", time.Second*39)),
		pipeline.NamedStep("build frontend", NoOpAction("buildfrontend", time.Minute)),
		pipeline.NamedStep("build documentation", NoOpAction("builddocs", time.Second*9)),
	)

	sw.Add(
		pipeline.NamedStep("test backend", NoOpAction("testbackend", time.Second*27)),
		pipeline.NamedStep("test frontend", NoOpAction("testfrontend", time.Second*32)),
	)

	sw.Add(
		pipeline.NamedStep("integration tests: sqlite", IntegrationTest("integrationtests_sqlite", time.Minute)),
		pipeline.NamedStep("integration tests: postgres", IntegrationTest("integrationtests_pg", time.Second*42)),
		pipeline.NamedStep("integration tests: mysql", IntegrationTest("integrationtests_mysql", time.Second*32)),
		pipeline.NamedStep("integration tests: mssql", IntegrationTest("integrationtests_mssql", time.Second*55)),
	)

	sw.Add(
		pipeline.NamedStep("package", NoOpAction("package", time.Second*13)),
		pipeline.NamedStep("build docker image", NoOpAction("builddocker", time.Second*44)),
	)

	sw.Add(
		pipeline.NamedStep("publish documentation", NoOpAction("publishdocs", time.Second*13)),
		pipeline.NamedStep("publish package", NoOpAction("publishpackage", time.Second*12)),
		pipeline.NamedStep("publish docker image", NoOpAction("publishdockerimage", time.Second*23)),
	)

	sw.Add(
		pipeline.NamedStep("notify slack", NoOpAction("notifyslack", time.Second*3)),
	)
}
