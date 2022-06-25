# Reactive Company 

## Introduction

This project is intended to demonstrate best practices for building 
a reactive web application with Spring 5 platform. It is a simple 
shopping application with some basic functionalities:
- listing blog posts/projects,
- adding new blog posts/projects,
- some filters for listings,
- rest requests for stream data.

There are some default data in database, so we don't
need to generate the new one for this web-service.

From the technical view, it's simple reactive (non-blocking) web 
service. In this release, Reactive-Company  is fully dockerized and 
can be benchmark-tested by
[wrk2](https://github.com/giltene/wrk2).

### What's in the box?

#### Application Components

- [Reactive-Company](src) - server with all the functionalities 
(including UI on port 8080)

#### Infrastructure Components

- [MongoDB](https://docs.mongodb.com/) - storage for all the data

## Adaptation

DeathStarBench is not a ready, standalone framework which we can run
as single component or import to the existing project. It is a kind
of prototype for benchmarking web applications. We don't have any
instruction how to adapt more services there, instead of that fact
we have three example services with description how to run and test
them. In Reactive-Company we provided the same instruction in this
README file below. We are also leaving a small note how we have
adapted Reactive-Company to DeathStarBench and made it able to be
benchmarked (stress-test):

1. Forked DeathStarBench for our purposes.
2. Downloaded DeathStarBench and Reactive-Company into separate directories.
3. Moved Reactive-Company to the DeathStarBench repository as a new service.
4. Changed some docker configuration to be able to run it on localhost.
5. Provided wrk2 in Reactive-Company as our test framework
6. Prepared some special request generators as .lua scripts.
7. Just tried to run application and wrk2 with our scripts.
8. Prepared documentation for new service.

## Running instructions

#### Pre-requirements

- Docker (you should be able to run it in swarm mode)
- Docker-compose
- luarocks (apt-get install luarocks)
- luasocket (luarocks install luasocket)

#### Spinning up the environment

First, spin up all the needed components

```bash
./docker-swarm.sh
```

On linux, it may be necessary to run these commands with ```sudo```

#### Workload generation

Move to the directory, where the benchmarking tool is and build it

```bash
cd wrk2
make
```

Now you can run some of the example tests:
1. Adding new blog post

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/reactive-company/add_blogpost.lua http://localhost:8080/blogposts/ -R <reqs-per-sec>
```

2. Getting all blog posts

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/blogposts/ -R <reqs-per-sec>
```

3. Getting all projects

```bash
./wrk -t <num-threads> -c <num-conns> -d <duration> -L -s ./scripts/reactive-company/get_all_projects.lua http://localhost:8080/projects/ -R <reqs-per-sec>
```

#### Stopping service

Exiting this application is not as intuitive as in other services,
where we are able to just call ```docker-compose down```. Because
this service is running in swarm mode, it is recommended to stop
it by executing this command

```bash
docker swarm leave --force
```

On linux, it may be necessary to run these commands with ```sudo```
