FROM ubuntu:18.04

ARG OPENTRACING_CPP_VERSION=v1.4.0
ARG JAEGER_VERSION=v0.4.1

COPY . /example

RUN set -x \
  && apt-get update \
  && apt-get install --no-install-recommends --no-install-suggests -y \
              build-essential \
              gettext \
              cmake \
              git \
              gnupg2 \
              software-properties-common \
              curl \
              ca-certificates \
              wget \
              lua5.1 lua5.1-dev \
### Build opentracing-cpp
  && cd / \
  && git clone -b $OPENTRACING_CPP_VERSION https://github.com/opentracing/opentracing-cpp.git \
  && cd opentracing-cpp \
  && mkdir .build && cd .build \
  && cmake -DBUILD_STATIC_LIBS=OFF -DBUILD_TESTING=OFF .. \
  && make && make install \
### Build bridge tracer
  && cd / \
  && git clone https://github.com/opentracing/lua-bridge-tracer.git \
  && cd lua-bridge-tracer \
  && mkdir .build && cd .build \
  && cmake .. \
  && make && make install \
### Install luvit
  && cd / \
  && curl -L https://github.com/luvit/lit/raw/master/get-lit.sh | sh \
### Install tracers
  && wget https://github.com/jaegertracing/jaeger-client-cpp/releases/download/${JAEGER_VERSION}/libjaegertracing_plugin.linux_amd64.so -O /usr/local/lib/libjaegertracing_plugin.so \
### Run ldconfig
  && ldconfig

EXPOSE 8080

ENTRYPOINT ["/luvit", "/example/server.lua"]
