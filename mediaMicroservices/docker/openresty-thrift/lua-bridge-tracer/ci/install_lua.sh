#!/bin/bash

set -e
LUA_VERSION="${LUA_VERSION:-5.3.4}"
cd /
curl -R -O http://www.lua.org/ftp/lua-$LUA_VERSION.tar.gz
tar zxf lua-$LUA_VERSION.tar.gz
cd lua-$LUA_VERSION
make linux install
