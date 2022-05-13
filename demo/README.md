# Demo pipelines

These demo pipelines are available to demosntrate a valuable structure of a pipeline.

Running a demo pipeline with the Shipwright CLI utility (requires `mage`):

1. Clone and `cd` into the Shipwright project: `git clone git@github.com:grafana/shipwright.git && cd shipwright`
2. Compile the shipwright CLI: `mage build`
3. Run the pipeline: `./bin/shipwright ./demo/{demo}`.

Running a demo pipeline without the Shipwright CLI:

1. Clone and `cd` into the Shipwright project: `git clone git@github.com:grafana/shipwright.git && cd shipwright`
2. Run the pipeline: `PIPELINE=./demo/{pipeline} go run -path=$PIPELINE $PIPELINE`

## [`./basic`](./basic)

This basic pipeline creates a single pipeline which runs many common steps that most projects might have.

### Features

- [ ] Background steps
- [ ] Caching
- [ ] Registering a new client
- [ ] Environment Variables
- [x] Event filters / triggers
- [x] Event variables
- [ ] Logging
- [ ] Multiple pipelines
- [ ] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [x] Running steps in sequence
- [ ] Running steps in parallel
- [x] Secrets
- [ ] Sub-pipelines
- [x] Step Arguments
- [x] State Management / Sharing data between steps
- [ ] Tracing

## [`./complex`](./complex)

This more complex pipeline creates a single pipeline which runs many steps with logs and timeouts. It has many steps and demonstrates a maintainable approach to writing large pipelines.

### Features

- [ ] Background steps
- [ ] Caching
- [ ] Registering a new client
- [ ] Environment Variables
- [ ] Event filters / triggers
- [ ] Event variables
- [ ] Logging
- [ ] Multiple pipelines
- [ ] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [x] Running steps in sequence
- [x] Running steps in parallel
- [ ] Secrets
- [ ] Sub-pipelines
- [x] Step Arguments
- [ ] State Management / Sharing data between steps
- [ ] Tracing

## [`./multi`](./multi)

This program creates multiple pipelines which run in sequence; one of which will only run if a commit on the main branch was tagged with a string starting with `v`.

### Features

- [ ] Background steps
- [ ] Caching
- [ ] Registering a new client
- [ ] Environment Variables
- [x] Event filters / triggers
- [x] Event variables
- [ ] Logging
- [x] Multiple pipelines
- [x] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [x] Running steps in sequence
- [x] Running steps in parallel
- [x] Secrets
- [ ] Sub-pipelines
- [x] Step Arguments
- [ ] State Management / Sharing data between steps
- [ ] Tracing

## [`./multi-sub`](./multi-sub)

This program creates multiple pipelines which run in sequence, and one pipeline which runs independently of the others.

### Features

- [ ] Background steps
- [ ] Caching
- [ ] Registering a new client
- [ ] Environment Variables
- [x] Event filters / triggers
- [ ] Event variables
- [ ] Logging
- [x] Multiple pipelines
- [x] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [x] Running steps in sequence
- [x] Running steps in parallel
- [x] Secrets
- [x] Sub-pipelines
- [x] Step Arguments
- [ ] State Management / Sharing data between steps
- [ ] Tracing

## [`./state`](./sub)

This program creates a very simple pipeline which demostrates setting and getting from the state.

### Features

- [ ] Background steps
- [ ] Caching
- [ ] Registering a new client
- [ ] Environment Variables
- [ ] Event filters / triggers
- [ ] Event variables
- [ ] Logging
- [ ] Multiple pipelines
- [ ] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [x] Running steps in sequence
- [ ] Running steps in parallel
- [ ] Secrets
- [ ] Sub-pipelines
- [x] Step Arguments
- [x] State Management / Sharing data between steps
- [ ] Tracing

## [`./custom-client`](./custom-client)

This program creates a very simple pipeline but with a new custom client available.

Use the `-client=my-custom-client` to run the pipeline with the custom client.

### Features

- [ ] Background steps
- [ ] Caching
- [x] Registering a new client
- [ ] Environment Variables
- [ ] Event filters / triggers
- [ ] Event variables
- [ ] Logging
- [ ] Multiple pipelines
- [ ] Running pipelines in sequence
- [ ] Running pipelines in parallel
- [ ] Running steps in sequence
- [ ] Running steps in parallel
- [ ] Secrets
- [ ] Sub-pipelines
- [ ] Step Arguments
- [ ] State Management / Sharing data between steps
- [ ] Tracing
