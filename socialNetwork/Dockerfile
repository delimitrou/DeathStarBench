FROM yg397/thrift-microservice-deps:xenial AS builder

ARG LIB_REDIS_PLUS_PLUS_VERSION=1.2.3

# Apply patch and re-install Redis plus plus
RUN cd /tmp/redis-plus-plus\
&& sed -i '/Transaction transaction/i\\    ShardsPool* get_shards_pool(){\n        return &_pool;\n    }\n' \
   src/sw/redis++/redis_cluster.h \
&& cmake -DREDIS_PLUS_PLUS_USE_TLS=ON . \
&& make -j$(nproc) \
&& make install

COPY ./ /social-network-microservices
RUN cd /social-network-microservices \
    && mkdir -p build \
    && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Debug .. \
    && make -j$(nproc) \
    && make install

FROM ubuntu:16.04

# Copy compiled C++ binaries and dependencies
COPY --from=builder /usr/local/bin/* /usr/local/bin/
COPY --from=builder /usr/local/lib/* /usr/local/lib/

# Install system dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        openssl \
        ca-certificates \
        libsasl2-2 \
        libmemcached11 \
        libmemcachedutil2 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /social-network-microservices
