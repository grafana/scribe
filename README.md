# scribe

Scribe is a framework for [Dagger](https://github.com/dagger/dagger) for writing flexible CI pipelines in Go that have consistent behavior when ran locally or in a CI server.

Write your pipeline once, run it locally and produce the config for your CI provider from the same code.

## Status

This is still in beta. Expect breaking changes and incomplete features.

## Why?

With Scribe you can:

- Write pipelines in a bottom-up declarative framework.
- Write pipeline steps in Go instead of shell scripts / bash.
- Run pipelines locally for testing using [Dagger](https://github.com/dagger/dagger).
- Generate configurations for existing CI providers.
- Use Go features to make complex pipelines easier to develop and maintain.

## Getting started

Check out the [`demos`](/demos) folder for some inspiration.
