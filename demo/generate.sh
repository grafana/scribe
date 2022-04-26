#!/usr/bin/env bash

for demo in ./demo/* ; do
    if [ -d "$demo" ]; then
      go run $demo -path=$demo -mode=drone -log-level=debug > $demo/gen_drone.yml
    fi
done
