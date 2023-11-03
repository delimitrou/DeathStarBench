FROM ubuntu:22.04 as builder

WORKDIR /workspace

ARG LUA_JIT_VERSION=2.1

RUN apt-get update -y && \
    apt-get install -y git libssl-dev libz-dev luarocks make && \
    wget https://github.com/LuaJIT/LuaJIT/archive/refs/tags/v${LUA_JIT_VERSION}.ROLLING.tar.gz && \
    tar -zxf v${LUA_JIT_VERSION}.ROLLING.tar.gz

RUN mkdir -p /build/deps/luajit && \
    cp -r LuaJIT-${LUA_JIT_VERSION}.ROLLING/* /build/deps/luajit

WORKDIR /build

COPY src/ src/
COPY Makefile Makefile

RUN make clean && make

FROM ubuntu:22.04

COPY --from=builder /build/wrk /usr/local/bin/wrk

RUN apt-get update -y && \
    apt-get install -y gcc luarocks && \
    luarocks install luasocket

CMD ["/bin/bash"]