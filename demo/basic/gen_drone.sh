#!/usr/bin/env bash

go run ./demo/basic -path=./demo/basic -mode=drone -log-level=info -version=latest > ./demo/basic/gen_drone.yml
