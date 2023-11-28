FROM ubuntu:16.04 as builder

ARG LIB_MONGOC_VERSION=1.15.0
ARG LIB_THRIFT_VERSION=0.12.0
ARG LIB_JSON_VERSION=3.6.1
ARG LIB_CPP_JWT_VERSION=1.1.1
ARG LIB_YAML_VERSION=0.6.2
ARG LIB_OPENTRACING_VERSION=1.6.0
ARG LIB_JAEGER_VERSION=0.9.0
ARG LIB_AMQP_CPP_VERSION=4.1.4
ARG LIB_SIMPLEAMQPCLIENT_VERSION=2.4.0
ARG LIB_CPP_REDIS_VERSION=4.3.1
ARG LIB_HIREDIS_VERSION=1.0.0
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
    && wget https://github.com/mongodb/mongo-c-driver/releases/download/${LIB_MONGOC_VERSION}/mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
    && tar -zxf mongo-c-driver-${LIB_MONGOC_VERSION}.tar.gz \
    && cd mongo-c-driver-${LIB_MONGOC_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DENABLE_TESTS=OFF -DENABLE_EXAMPLES=OFF -DENABLE_AUTOMATIC_INIT_AND_CLEANUP=OFF .. \
    && make -j$(nproc) \
    && make install

# Install lib-thrift
RUN cd /tmp \
    && wget -O thrift-${LIB_THRIFT_VERSION}.tar.gz https://github.com/apache/thrift/archive/v${LIB_THRIFT_VERSION}.tar.gz \
    && tar -zxf thrift-${LIB_THRIFT_VERSION}.tar.gz \
    && cd thrift-${LIB_THRIFT_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF -DBUILD_PYTHON=OFF -DBUILD_C_GLIB=OFF -DBUILD_EXAMPLES=OFF -DBUILD_EXAMPLES=OFF -DBUILD_TUTORIALS=OFF .. \
    && make -j$(nproc) \
    && make install

# Install /nlohmann/json
RUN cd /tmp \
    && wget -O json-${LIB_JSON_VERSION}.tar.gz https://github.com/nlohmann/json/archive/v${LIB_JSON_VERSION}.tar.gz \
    && tar -zxf json-${LIB_JSON_VERSION}.tar.gz \
    && cd json-${LIB_JSON_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_TESTING=OFF .. \
    && make -j$(nproc) \
    && make install \
# Install jwt
    && cd /tmp \
    && wget -O cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz https://github.com/arun11299/cpp-jwt/archive/v${LIB_CPP_JWT_VERSION}.tar.gz \
    && tar -zxf cpp-jwt-${LIB_CPP_JWT_VERSION}.tar.gz \
    && cd cpp-jwt-${LIB_CPP_JWT_VERSION} \
    && cp -R include/jwt /usr/local/include \
    # use the dependency in /usr/local/include instead of in jwt/json
    && rm -rf /usr/local/include/jwt/json \
    && sed -i 's/\#include \"jwt\/json\/json.hpp\"/\#include \<nlohmann\/json\.hpp\>/g' /usr/local/include/jwt/jwt.hpp \
# Install yaml-cpp
    && cd /tmp \
    && wget -O yaml-cpp-${LIB_YAML_VERSION}.tar.gz https://github.com/jbeder/yaml-cpp/archive/yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
    && tar -zxf yaml-cpp-${LIB_YAML_VERSION}.tar.gz \
    && cd yaml-cpp-yaml-cpp-${LIB_YAML_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DYAML_CPP_BUILD_TESTS=OFF .. \
    && make -j$(nproc) \
    && make install

# Install opentracing-cpp
RUN cd /tmp \
    && wget -O opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz https://github.com/opentracing/opentracing-cpp/archive/v${LIB_OPENTRACING_VERSION}.tar.gz \
    && tar -zxf opentracing-cpp-${LIB_OPENTRACING_VERSION}.tar.gz \
    && cd opentracing-cpp-${LIB_OPENTRACING_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DBUILD_TESTING=OFF -DBUILD_MOCKTRACER=OFF .. \
    && make -j$(nproc) \
    && make install \
# Install jaeger-client-cpp
    && cd /tmp \
    && wget -O jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz https://github.com/jaegertracing/jaeger-client-cpp/archive/v${LIB_JAEGER_VERSION}.tar.gz \
    && tar -zxf jaeger-client-cpp-${LIB_JAEGER_VERSION}.tar.gz \
    && cd jaeger-client-cpp-${LIB_JAEGER_VERSION} \
    && mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_CXX_FLAGS="-fPIC" -DHUNTER_ENABLED=OFF -DBUILD_TESTING=OFF -DJAEGERTRACING_WITH_YAML_CPP=ON -DJAEGERTRACING_BUILD_EXAMPLES=OFF .. \
    && make -j$(nproc) \
    && make install

# Install AMQP-CPP
RUN cd /tmp \
    && git clone https://github.com/CopernicaMarketingSoftware/AMQP-CPP.git \
    && cd AMQP-CPP \
    && git checkout v${LIB_AMQP_CPP_VERSION} \
    && mkdir cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release -DAMQP-CPP_BUILD_SHARED=ON -DAMQP-CPP_LINUX_TCP=ON \
    && make -j$(nproc) \
    && make install \
# Install SimpleAmqpClient
    && cd /tmp \
    && git clone https://github.com/alanxz/SimpleAmqpClient.git \
    && cd SimpleAmqpClient \
    && git checkout v${LIB_SIMPLEAMQPCLIENT_VERSION} \
    && mkdir cmake-build \
    && cd cmake-build \
    && cmake .. -DCMAKE_BUILD_TYPE=Release \
    && make -j$(nproc) \
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
    && make -j$(nproc) \
    && make install \
# Install Hiredis
    && cd /tmp \
    && git clone https://github.com/redis/hiredis.git \
    && cd hiredis \
    && git checkout v${LIB_HIREDIS_VERSION} \
    && make -j$(nproc) USE_SSL=ON \
    && make USE_SSL=1 install \
# Install Redis plus plus
    && cd /tmp \
    && git clone https://github.com/sewenew/redis-plus-plus.git \
    && cd redis-plus-plus \
    && git checkout ${LIB_REDIS_PLUS_PLUS_VERSION} \
    && sed -i '/Transaction transaction/i\\    ShardsPool* get_shards_pool(){\n        return &_pool;\n    }\n' src/sw/redis++/redis_cluster.h \
    && cmake -DREDIS_PLUS_PLUS_USE_TLS=ON . \
    && make -j$(nproc) \
    && make install

RUN ldconfig -v

WORKDIR /workspace

COPY src/ src/
COPY cmake/ cmake/
COPY gen-cpp/ gen-cpp/
COPY third_party/ third_party/
COPY CMakeLists.txt CMakeLists.txt

RUN mkdir -p cmake-build \
    && cd cmake-build \
    && cmake -DCMAKE_BUILD_TYPE=Release .. \
    && make -j$(nproc) \
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
