#!/bin/bash

set -e

LUAROCKS_VERSION=2.4.4
cd /
wget http://luarocks.github.io/luarocks/releases/luarocks-${LUAROCKS_VERSION}.tar.gz
tar zxf luarocks-${LUAROCKS_VERSION}.tar.gz
cd luarocks-${LUAROCKS_VERSION}
./configure
make build
make install

luarocks install busted
luarocks install json-lua
