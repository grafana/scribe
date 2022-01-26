version=$(shell git describe --tags --dirty --always)

default: build

build:
	go build \
		-ldflags \
		"-X main.Version=$(version)" \
		-o bin/shipwright ./plumbing/cmd

test:
	go test ./...
