#!/bin/bash

set -e
OPENTRACING_VERSION="${OPENTRACING_VERSION:-1.5.0}"
BUILD_DIR=/
pushd "${BUILD_DIR}"
git clone -b v$OPENTRACING_VERSION https://github.com/opentracing/opentracing-cpp.git
cd opentracing-cpp
mkdir .build && cd .build
cmake -DCMAKE_BUILD_TYPE=Debug \
      -DBUILD_MOCKTRACER=ON \
      -DBUILD_TESTING=OFF \
      ..
make && make install
popd
