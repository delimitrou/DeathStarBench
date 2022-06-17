# ContosoCrafts (Steeltoe Edition)

## Introduction

Simple shopping application with some basic functionalities:
- listing products,
- rating single product, 
- adding product to cart,
- submitting order.

There are some default home-products in database, so we don't
need to generate data for this web-service.

From the technical view, it's simple distributed system based on
microservices concept. In this release, ContosoCrafts is fully
dockerized and can be benchmark-tested by 
[wrk2](https://github.com/giltene/wrk2).

### What's in the box?

#### Application Components

- [Contoso Website](src/ContosoCrafts.WebSite)
- [Products API](src/ContosoCrafts.ProductsApi)
- [Checkout Processor](src/ContosoCrafts.CheckoutProcessor)

#### Infrastructure Components

- [Redis](https://redis.io/) - State store
- [RabbitMQ](https://www.rabbitmq.com/) - Message Broker
- [Zipkin](https://zipkin.io/) - Distributed tracing
- [MongoDB](https://docs.mongodb.com/) - Products data
- [Fluent Bit](https://fluentbit.io/) - Log forwarder
- [Seq](https://datalust.co/seq) - Log Aggregator

## Adaptation

DeathStarBench is not a ready, standalone framework which we can run
as single component or import to the existing project. It is a kind
of prototype for benchmarking web applications. We don't have any
instruction how to adapt more services there, instead of that fact
we have three example services with description how to run and test
them. In ContosoCrafts we provided the same instruction in this
README file below. We are also leaving a small note how we have
adapted ContosoCrafts to DeathStarBench and made it able to be
benchmarked (stress-test):

1. Forked DeathStarBench for our purposes.
2. Downloaded DeathStarBench and ContosoCrafts into separate directories.
3. Moved ContosoCrafts to the DeathStarBench repository as a new service.
4. Provided wrk2 in ContosoCrafts as our test framework
5. Prepared some special request generators as .lua scripts.
6. Just tried to run application and wrk2 with our scripts.
7. Prepared documentation for new service.

## Running

#### Pre-requirements

- Docker
- Docker-compose
- luarocks (apt-get install luarocks)
- luasocket (luarocks install luasocket)

#### Spinning up the environment

First, spin up the supporting infrastructure components

```bash
> docker-compose -f docker-compose-infra.yml up -d
```

Next, launch the application containers and sidecars.

```bash
> docker-compose up -d
```

On linux, it may be necessary to run these commands with ```sudo```

#### Workload generation

Move to the directory, where the benchmarking tool is and build it

```bash
cd wrk2
make
```

Now you can run some of the example tests:
1. Getting single random products (NOTE: fill the params in command)

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/contosoCrafts/get_singles.lua http://localhost:9090/Products/Index= -R <reqs-per-sec>
```

2. Evaluation of random products

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/contosoCrafts/rate_product.lua http://localhost:9090/Products/ -R <reqs-per-sec>
```

3. Getting all products

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/contosoCrafts/get_all.lua http://localhost:9090/Products/ -R <reqs-per-sec>
```

1. Mixed scenario 

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/contosoCrafts/mix_scenario.lua http://localhost:9090/Products/ -R <reqs-per-sec>
```