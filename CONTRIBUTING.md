# Development guide

## Layout

```
.
|─./shipwright.go
├── ci
├── demo
├── {package}
│   └── x
└── plumbing
    ├── clients
    ├── cmd
    │   └── commands
    ├── pipeline
    ├── plog
    ├── schemas
    └── {x}util
```

| directory / format   | description                                                                                                                                                                     |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `./shipwright*.go`   | Defines the `Client` interface and wrapper types that Pipeline developers use to create pipelines.                                                                              |
| `./ci`               | Shipwright pipeline that tests / builds this repository.                                                                                                                        |
| `./demo`             | Demo pipelines. Each sub-directory should be a separate pipeline that introduces a new / separate concept.                                                                      |
| `./{package}         | Represents a Go package that should contain only definitions for **Steps** or **Actions** for use in Shipwright pipelines.                                                      |
| `./{package}/x       | The small, unit-testable functions that power the actions used in `{package}`.                                                                                                  |
| `./plumbing          | The packages that power the pipeline logic including asyncronous / goroutine handling and client code.                                                                          |
| `./plumbing/clients  | The Clients that satisfy the `Client` interface. These Clients can run the pipeline in an environment of some kind, or can generate configuration that represents the pipeline. |
| `./plumbing/cmd      | The `main` package and commands that make up the `shipwright` binary.                                                                                                           |
| `./plumbing/pipeline | The types that make up a Pipeline, regardless of client. Primarily `Step` and `StepAction`.                                                                                     |
| `./plumbing/plog     | The Logger that is used in Shipwright Clients.                                                                                                                                  |
| `./plumbing/schemas  | Strictly contains types that represent third-party configuration schemas that Clients will use for generation. (TODO: Maybe those schemas should live next to the clients?      |
| `./plumbing/{x}util  | Specific utility packages that help with {x}. For example, {sync}util helps with using the {sync} package.                                                                      |

Important notes:

- Try to limit the amount of non-Step/StepAction logic in a `./{package}` to a minimum and delegate that logic to the `x` sub-package.
  - For example, the `Docker` client builds a Go binary of the requested pipeline, but does not perform this action in the pipeline as a Step, so it uses the `pkg.grafana.com/shipwright/v1/golang/x` package.

## Style guide

### Global style suggestions

- Prefer using standard library packages over third party ones.
  - `flag`, `os/exec`

## Command-line arguments

There are two places where command-line arguments are parsed:

1. In the `shipwright.New` function, which is the first function called in a pipeline.
2. In the `plumbing/cmd` package for parsing options supplied in the `shipwright` command.
