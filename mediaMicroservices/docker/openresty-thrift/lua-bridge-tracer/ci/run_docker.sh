#!/bin/bash

set -e

BUILD_IMAGE=lua-bridge-tracer
docker image inspect "$BUILD_IMAGE" &> /dev/null || {
  docker build -t "$BUILD_IMAGE" ci
}

if [ -n "$1" ]; then
  docker run -v "$PWD":/src -w /src -it "$BUILD_IMAGE" /bin/bash -lc "$1"
else
  docker run -v "$PWD":/src -w /src --privileged -it "$BUILD_IMAGE" /bin/bash -l
fi

