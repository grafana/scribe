# scribe

Scribe is a framework to write flexible CI pipelines in Go that have consistent behavior when ran locally or in a CI server.

## Status

This is still in beta. Expect breaking changes.

## Why?

Continuous integration and delivery are built on flaky foundations.

Configuration-based pipeline tooling has been prevalent all over open source software. It works in all languages, it's mostly platform agnostic, and offers flexibility with `bash`. However, it does not scale with several people, and its problems are seen quickly.

- Development and testing strictly involves trial and error.
- Automated testing is non-existent.
- Development tools are limited to defining configuration schemas, which provide a shallow understanding of what each key and value will do.
- Configuration languages have nuanced syntax which is typically a lot different than standard programming languages.
- Dependencies / dependency management is often not supported as it's just not a typical configuration language (yaml, json) construct.
- Lack of debugging means large / extensive pipelines are flaky and make it difficult to diagnose issues.
  - Problems like these are demoralizing and often lead to neglect; no one wants to address issues like these because they're so difficult to debug.
- Providers are incentivized to keep their YAML from being ran on other platforms. Many providers probably not like it if I could run a pipeline made for their platform inside a competitor's.

These problems lead to the development of this framework, **scribe**.

The idea behind `scribe` is that it is not an application, but a library. There is no server. Users should, instead of defining an amalgamation of `yaml/json/toml/whatever` and `bash`, define their build, package, and release processes programmatically. This opens up a whole world of possibilities, like:

- Writing unit and integration tests for your build pipeline.
- Reusing and sharing build, package, and deployment definitions.
- Creating reasonable packages and libraries for developing new pipelines.
- Improved visualization by allowing pipelines to define metrics and traces.
- Improved validation by having steps define what they expect in order to run, and what they provide to other steps.

## Glossary

- **Pipeline**: A pipeline is a generic sequence of steps. A pipeline can be a set of steps to build an application, or it can define how to take an artifact, package it, and push it to a package repository.
  - **Action**: A pipeline action is a single reusable component in a pipeline. Actions can have arguments and define outputs.
  - **Source**: A pipeline source defines what causes a pipeline to begin. For typical continuous integration builds, this source would be a commit or a tag. For a release pipeline, this could be a NATS message, a Google Cloud Pub/Sub message that says an artifact is available, or it could be another pipeline.
- **Artifact**: The tangible, end-result of a pipeline or step. Not all pipelines produce artifacts.
  - An example of an artifact would be a compiled binary or a docker image.
- **Target**: A target is a software release destination. It is the final place that an artifact is sent before it is used to serve user requests.

## Running Locally / testing

- Compile the Scribe utility: `mage build`
- Run the local pipeline in the current shell: `./bin/scribe-mode=cli ./ci`
- Run the local pipeline in docker: `./bin/scribe -mode=docker ./ci`
- Generate the drone: `./bin/scribe -mode=drone ./ci`
- Generate the drone and write it to a file: `./bin/scribe -mode=drone ./ci > .drone.yml`

## How does it work?

The main idea behind `scribe` is that it defers what is typically considered server logic into the client / pipeline definitions and library.

### Clients

Scribe clients are used in the pipelines themselves. All pipelines are programs, and thus have a `func main()`.

There are currently three planned Clients, which are selected using the `-mode` CLI argument.

1. `cli` - Runs the pipeline locally, attempting to emulate what would normally be executed in a standard CI pipeline.
2. `drone` - Generates a Drone configuration that matches the specified pipeline.
3. `docker` - Similar to `run`, but instead runs the pipeline using the Docker CLI

#### Run mode `docker`

The `docker` run mode will run each pipeline in a Docker image the same way that it would likely run in a CI platform. It's almost a combination of the `cli` and `drone` run modes.

Each step defined must have an image. For steps without defined images, the scribe will be used.

When running in docker mode, the pipeline is compiled and then mounted as volume in the docker container. The compiled pipeline is used as the docker command for that step.

## Writing a Pipeline

1. Every pipeline is a program and should have a `package main` and a `func main`.
2. Every pipeline must have a form of `sw.New(...)` or `sw.NewMulti(...)` to produce the scribe object.

   - Steps are then added to that object to create a pipeline.

3. It is recommended to create a Go workspace for your CI pipeline with `go work init {directory}`.

   - This will keep the larger and irrelevant modules like `docker` out of your project.

### Examples

To view examples of pipelines, visit the [demo](./demo) folder. These demos are used in our automated tests.

## FAQ

- **Why use Go and not `JavaScript/TypeScript/Python/Java`?**

We use Go pretty ubiquitously at Grafana, especially in our server code. Go also allows you to easily compile a static binary for Linux from any platform which helps a lot with the portability of Scribe, especially in Docker mode.

- **Will there be support for any other languages?**

Given the current design, it would be very difficult and there are no concrete plans to do that yet.

- **What clients are available?**

- `cli`, which runs the pipeline in the current shell.
- `docker`, which runs the pipeline using the docker daemon (configured via the Docker environment variables).
- `drone`, which produces a .drone.yml file in the standard output stream (`stdout`) that will run the pipeline in Drone.

The current list of clients can always be obtained using the `scribe -help` command.

- **How can I use unsupported clients or make my own?**

Because Scribe is simply a package and your pipeline is a program, you can add a client you have made yourself in your pipeline.

In the `init` function of the pipeline, simply register your client and it should be available for use. For a demonstration, see [`./demo/custom-client`](./demo/custom-client).
