FROM yg397/thrift-microservice-deps:xenial

ARG NUM_CPUS=40

COPY ./ /social-network-microservices
RUN cd /social-network-microservices \
    && mkdir -p build \
    && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Debug .. \
    && make -j${NUM_CPUS} \
    && make install

WORKDIR /social-network-microservices