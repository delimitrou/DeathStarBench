FROM yg397/thrift-microservice-deps:xenial

ARG NUM_CPUS=40

COPY ./ /media-microservices
RUN cd /media-microservices \
    && mkdir -p build \
    && cd build \
    && cmake .. \
    && make -j${NUM_CPUS} \
    && make install

WORKDIR /media-microservices