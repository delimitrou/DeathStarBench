#!/bin/bash

set -e

function setup_asan_flags()
{
  export CFLAGS="$CFLAGS -Werror -fno-omit-frame-pointer -fsanitize=address"
  export CXXFLAGS="$CXXFLAGS -Werror -fno-omit-frame-pointer -fsanitize=address"
  export LDFLAGS="$LDFLAGS -Werror -fno-omit-frame-pointer -fsanitize=address"
}

export MOCKTRACER=/usr/local/lib/libopentracing_mocktracer.so

function run_lua_test()
{
  setup_asan_flags
  ./ci/install_opentracing.sh
  ./ci/install_lua.sh
  ./ci/install_rocks.sh
  SRC_DIR=`pwd`
  mkdir /build && cd /build
  cmake -DCMAKE_BUILD_TYPE=Debug $SRC_DIR
  make && make install
  ldconfig
  cd $SRC_DIR
  LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libasan.so.4 busted test/tracer.lua
}

if [[ "$1" == "test-5.3" ]]; then
  export LUA_VERSION=5.3.4
  run_lua_test
  exit 0
elif [[ "$1" == "test-5.2" ]]; then
  export LUA_VERSION=5.2.4
  run_lua_test
  exit 0
elif [[ "$1" == "test-5.1" ]]; then
  export LUA_VERSION=5.1.5
  run_lua_test
  exit 0
elif [[ "$1" == "coverage" ]]; then
  export LUA_VERSION=5.1.5
  ./ci/install_opentracing.sh
  ./ci/install_lua.sh
  ./ci/install_rocks.sh
  SRC_DIR=`pwd`
  mkdir /build && cd /build
  cmake -DCMAKE_BUILD_TYPE=Debug \
    -DCMAKE_CXX_FLAGS="-fprofile-arcs -ftest-coverage -fPIC -O0" \
    $SRC_DIR
  make && make install
  ldconfig
  cd $SRC_DIR
  busted test/tracer.lua
  cd /build/CMakeFiles/opentracing_bridge_tracer.dir/src
  gcovr -r $SRC_DIR/src . --html --html-details -o coverage.html
  mkdir /coverage
  cp *.html /coverage/
fi
