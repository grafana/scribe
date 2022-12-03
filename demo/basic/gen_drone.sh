#!/usr/bin/env bash

go run ./demo/basic -path=./demo/basic -client=drone -log-level=info -version=latest > ./demo/basic/gen_drone.yml
