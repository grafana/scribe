#!/usr/bin/env bash

for demo in ./demo/* ; do
    if [ -d "$demo" ]; then
      echo "go run $demo -path=$demo -mode=drone > $demo/gen_drone.yml"
      go run $demo -path=$demo -mode=drone > $demo/gen_drone.yml
    fi
done
