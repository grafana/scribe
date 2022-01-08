# Development guide

## Requirements

- Prefer using standard library packages over third party ones.
  - `log`, `flag`, `os/exec`

## Command-line arguments

There are two places where command-line arguments are parsed:

1. In the `shipwright.New` function, which is the first function called in a pipeline.
2. In the `plumbing/cmd` package for parsing options supplied in the `shipwright` command.
