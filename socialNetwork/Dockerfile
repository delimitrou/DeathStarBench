FROM ubuntu:16.04 as builder

ARG LIB_MONGOC_VERSION=1.15.0
ARG LIB_THRIFT_VERSION=0.12.0
ARG LIB_JSON_VERSION=3.6.1
ARG LIB_CPP_JWT_VERSION=1.2
ARG LIB_YAML_VERSION=0.6.2
ARG LIB_OPENTRACING_VERSION=1.6.0
ARG LIB_JAEGER_VERSION=0.9.0
ARG LIB_AMQP_CPP_VERSION=4.1.4
ARG LIB_SIMPLEAMQPCLIENT_VERSION=2.4.0
ARG LIB_CPP_REDIS_VERSION=4.3.1
ARG LIB_HIREDIS_VERSION=1.0.2
ARG LIB_REDIS_PLUS_PLUS_VERSION=1.2.3

ENV LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:/usr/local/lib

RUN apt-get update \
    && apt-get install -y \
        automake \
        bison \
        ca-certificates \
        cmake \
        flex \
        git \
        g++ \
        libboost-all-dev \
        libevent-dev \
        libmemcached-dev \
        librabbitmq-dev \
        libssl-dev \
        libtool \
        make \
        pkg-config \
        wget

# Install mongo-c-driver
RUN cd /tmp \
    && wget -O mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz https://github.com/mongodb/mongo-c-driver/releases/download/${LIB_MONGOC_VERSION}/mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
    && tar -zxf mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
    && cd mongo-c-driver-${LIB_MONGOC_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DENABLE_TESTS=OFF -DENABLE_EXAMPLES=OFF -DENABLE_AUTOMATIC_INIT_AND_CLEANUP=OFF .. \
    && make \
    && make install

# Install lib-thrift
RUN cd /tmp \
    && wget -O thrift-${LIB_THRIFT_VERSION}.tar.gz https://github.com/apache/thrift/archive/refs/tags/v${LIB_THRIFT_VERSION}.tar.gz \
    && tar -zxf thrift-${LIB_THRIFT_VERSION}.tar.gz \
    && cd thrift-${LIB_THRIFT_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF -DBUILD_PYTHON=OFF -DBUILD_C_GLIB=OFF -DBUILD_EXAMPLES=OFF -DBUILD_EXAMPLES=OFF -DBUILD_TUTORIALS=OFF .. \
    && make \
    && make install

# Install /nlohmann/json
RUN cd /tmp \
    && wget -O json-${LIB_JSON_VERSION}.tar.gz https://github.com/nlohmann/json/archive/refs/tags/v${LIB_JSON_VERSION}.tar.gz \
    && tar -zxf json-${LIB_JSON_VERSION}.tar.gz \
    && cd json-${LIB_JSON_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF .. \
    && make \
    && make install \
# Install jwt
    && cd /tmp \
    && wget -O cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz https://github.com/arun11299/cpp-jwt/archive/refs/tags/v${LIB_CPP_JWT_VERSION}.tar.gz \
    && tar -zxf cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz \
    && cd cpp-jwt-${LIB_CPP_JWT_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCPP_JWT_USE_VENDORED_NLOHMANN_JSON=OFF -DCPP_JWT_BUILD_TESTS=OFF -DCPP_JWT_BUILD_EXAMPLES=OFF .. \
    && make \
    && make install \
# Install yaml-cpp
    && cd /tmp \
    && wget -O yaml-cpp-${LIB_YAML_VERSION}.tar.gz https://github.com/jbeder/yaml-cpp/archive/refs/tags/yaml-cpp-${LIB_YAML_VERSION}.tar.gz\
    && tar -zxf yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
    && cd yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DYAML_CPP_BUILD_TESTS=OFF .. \
    && make \
    && make install

# Install opentracing-cpp
RUN cd /tmp \
    && wget -O opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz https://github.com/opentracing/opentracing-cpp/archive/refs/tags/v${LIB_OPENTRACING_VERSION}.tar.gz \
    && tar -zxf opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz \
    && cd opentracing-cpp-${LIB_OPENTRACING_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DBUILD_TESTING=OFF -DBUILD_MOCKTRACER=OFF .. \
    && make \
    && make install \
# Install jaeger-client-cpp
    && cd /tmp \
    && wget -O jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz https://github.com/jaegertracing/jaeger-client-cpp/archive/refs/tags/v${LIB_JAEGER_VERSION}.tar.gz \
    && tar -zxf jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz \
    && cd jaeger-client-cpp-${LIB_JAEGER_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DHUNTER_ENABLED=OFF -DBUILD_TESTING=OFF -DJAEGERTRACING_WITH_YAML_CPP=ON -DJAEGERTRACING_BUILD_EXAMPLES=OFF .. \
    && make \
    && make install

# Install AMQP-CPP
RUN cd /tmp \
    && wget -O AMQP-CPP-${LIB_AMQP_CPP_VERSION}.tar.gz https://github.com/CopernicaMarketingSoftware/AMQP-CPP/archive/refs/tags/v${LIB_AMQP_CPP_VERSION}.tar.gz \
    && tar -zxf AMQP-CPP-${LIB_AMQP_CPP_VERSION}.tar.gz  \
    && cd AMQP-CPP-${LIB_AMQP_CPP_VERSION} \
    && mkdir cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release -DAMQP-CPP_BUILD_SHARED=ON -DAMQP-CPP_LINUX_TCP=ON \
    && make \
    && make install \
# Install SimpleAmqpClient
    && cd /tmp \
    && wget -O SimpleAmqpClient-${LIB_SIMPLEAMQPCLIENT_VERSION}.tar.gz https://github.com/alanxz/SimpleAmqpClient/archive/refs/tags/v${LIB_SIMPLEAMQPCLIENT_VERSION}.tar.gz \
    && tar -zxf SimpleAmqpClient-${LIB_SIMPLEAMQPCLIENT_VERSION}.tar.gz \
    && cd SimpleAmqpClient-${LIB_SIMPLEAMQPCLIENT_VERSION} \
    && mkdir cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release \
    && make \
    && make install

# Install cpp_redis
RUN cd /tmp \
    && git clone https://github.com/cpp-redis/cpp_redis.git \
    && cd cpp_redis \
    && git checkout ${LIB_CPP_REDIS_VERSION} \
    && git submodule init \
    && git submodule update \
    && mkdir cmake-build && \
    cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release \
    && make \
    && make install \
# Install Hiredis
    && cd /tmp \
    && wget -O hiredis-${LIB_HIREDIS_VERSION}.tar.gz https://github.com/redis/hiredis/archive/refs/tags/v${LIB_HIREDIS_VERSION}.tar.gz \
    && tar -zxf hiredis-${LIB_HIREDIS_VERSION}.tar.gz \
    && cd hiredis-${LIB_HIREDIS_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release -DENABLE_SSL=ON -DDISABLE_TESTS=ON \
    && make \
    && make install \
# Install Redis plus plus
    && cd /tmp \
    && wget -O redis-plus-plus-${LIB_REDIS_PLUS_PLUS_VERSION}.tar.gz https://github.com/sewenew/redis-plus-plus/archive/refs/tags/${LIB_REDIS_PLUS_PLUS_VERSION}.tar.gz \
    && tar -zxf redis-plus-plus-${LIB_REDIS_PLUS_PLUS_VERSION}.tar.gz \
    && cd redis-plus-plus-${LIB_REDIS_PLUS_PLUS_VERSION} \
    && sed -i '/Transaction transaction/i\\    ShardsPool* get_shards_pool(){\n        return &_pool;\n    }\n' src/sw/redis++/redis_cluster.h \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release -DREDIS_PLUS_PLUS_USE_TLS=ON -DREDIS_PLUS_PLUS_BUILD_TEST=OFF \
    && make \
    && make install

RUN ldconfig

WORKDIR /workspace

COPY src/ src/
COPY cmake/ cmake/
COPY gen-cpp/ gen-cpp/
COPY third_party/ third_party/
COPY CMakeLists.txt CMakeLists.txt

RUN mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release .. \
    && make \
    && make install

FROM ubuntu:16.04

# Copy compiled C++ binaries and dependencies
COPY --from=builder /usr/local/bin/ /usr/local/bin/
COPY --from=builder /usr/local/lib/ /usr/local/lib/

# Install system dependencies
RUN apt-get update \
    && apt-get install -y \
        openssl \
        ca-certificates \
        libsasl2-2 \
        libmemcached11 \
        libmemcachedutil2 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /social-network-microservices
