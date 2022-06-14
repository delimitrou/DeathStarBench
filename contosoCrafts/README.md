# ContosoCrafts (Steeltoe Edition)

## Spinning up the environment

First, spin up the supporting infrastructure components

```bash
> docker-compose -f docker-compose-infra.yml up -d
```

Next, launch the application containers and sidecars.

```bash
> docker-compose up -d
```

### Requirements

- Docker
- Visual Studio Code
- .NET Core SDK

## What's in the box

### Application Components

- [Contoso Website](src/ContosoCrafts.WebSite)
- [Products API](src/ContosoCrafts.ProductsApi)
- [Checkout Processor](src/ContosoCrafts.CheckoutProcessor)

### Infrastructure Components

- [Redis](https://redis.io/) - State store
- [RabbitMQ](https://www.rabbitmq.com/) - Message Broker
- [Zipkin](https://zipkin.io/) - Distributed tracing
- [MongoDB](https://docs.mongodb.com/) - Products data
- [Fluent Bit](https://fluentbit.io/) - Log forwarder
- [Seq](https://datalust.co/seq) - Log Aggregator
