FROM ubuntu:22.04 as builder

WORKDIR /workspace

RUN apt-get update -y && \
    apt-get install -y libssl-dev libz-dev luarocks make && \
    luarocks install luasocket

COPY src/ src/
COPY deps/ deps/
COPY Makefile Makefile

RUN make && make install

FROM ubuntu:22.04

COPY --from=builder /usr/local/bin/wrk /usr/local/bin/wrk
COPY --from=builder /usr/local/lib/lua /usr/local/lib/lua
COPY --from=builder /usr/local/lib/luarocks /usr/local/lib/luarocks
COPY --from=builder /usr/local/share/lua /usr/local/share/lua

CMD ["/bin/bash"]