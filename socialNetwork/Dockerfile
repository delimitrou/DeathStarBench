FROM yg397/thrift-microservice-deps:xenial AS builder

ARG LIB_JAEGER_VERSION=0.9.0
ARG LIB_OPENTRACING_VERSION=1.6.0
ARG LIB_REDIS_PLUS_PLUS_VERSION=1.2.3

# Install Redis plus plus with patch
RUN cd /tmp/redis-plus-plus \
    && sed -i '/Transaction transaction/i\\    ShardsPool* get_shards_pool(){\n        return &_pool;\n    }\n' \
        src/sw/redis++/redis_cluster.h \
    && cmake -DREDIS_PLUS_PLUS_USE_TLS=ON . \
    && make -j$(nproc) \
    && make install \
    # Updates opentracing-cpp
    && cd /tmp \
    && wget -O opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz https://github.com/opentracing/opentracing-cpp/archive/v${LIB_OPENTRACING_VERSION}.tar.gz \
    && tar -zxf opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz \
    && cd opentracing-cpp-${LIB_OPENTRACING_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DBUILD_TESTING=0 .. \
    && make -j$(nproc) \
    && make install \
    # Updates jaeger-client-cpp
    && cd /tmp \
    && wget -O jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz https://github.com/jaegertracing/jaeger-client-cpp/archive/v${LIB_JAEGER_VERSION}.tar.gz \
    && tar -zxf jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz \
    && cd jaeger-client-cpp-${LIB_JAEGER_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DHUNTER_ENABLED=0 -DBUILD_TESTING=0 -DJAEGERTRACING_WITH_YAML_CPP=1 -DJAEGERTRACING_BUILD_EXAMPLES=0 .. \
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
COPY --from=builder /usr/local/bin/ /usr/local/bin/
COPY --from=builder /usr/local/lib/ /usr/local/lib/

# Install system dependencies
RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
        openssl \
        ca-certificates \
        libsasl2-2 \
        libmemcached11 \
        libmemcachedutil2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /social-network-microservices
