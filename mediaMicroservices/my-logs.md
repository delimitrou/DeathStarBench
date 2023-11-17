# Initial setup

sudo apt install python3-pip

pip3 install aiohttp asyncio

sudo apt-get install libssl-dev

sudo apt-get install libz-dev

sudo apt-get install luarocks

sudo luarocks install luasocket

sudo apt install docker-compose

cd DeathStarBench/wrk2

make all

cd DeathStarBench/mediaMicroservices

# Running the media service application
Host actions are denoted as h: action and client actions are denoted as c: action

h: ./compile.sh (only needed if you change the code)

h: ./run.sh

h: sudo ./remote-perf-https-multicore-x86 1 24 (only needed if you want to save instruction counts per query)

c: update ip address in wrk2/scripts/media-microservices/compose-review-record.lua (only needed if you want to save instruction counts per query)

c: ./load.sh (Usage: $0 [-t \<num-threads>] [-c \<num-conns>] [-d \<duration ex: 2m>] [-i \<ip-addr>] [-r \<reqs-per-sec>] [-s \<save-results 1 or 0>] [-h])

# MongoDB problem Reproduction
2 Options:

1. Run the load script with the -s 1 flag to save the results to a file. Then plot the instruction count graph for core 6 or 7 for movie review mongodb and 20 or 21 for user review mongodb. The instruction count graph is an increasing function.

2. Run the load script without -s 1 flag. Check CPU utilization of user review and movie revies mongodb containers with docker stats. The CPU utilization increases with time.

We believe the problem is because of their database schema. For example their user_review schema is like this:
```{
    "user_id": "string",
    "reviews": [
        {
            "review_id": "string",
            "timestamp": "string"
        },
        {
            "review_id": "string",
            "timestamp": "string"
        }...
    ]
}
```
With each review, this object size increases and it has to copy the whole array. Here is a pointer for mongodb blog post about this problem: https://www.mongodb.com/docs/atlas/schema-suggestions/avoid-unbounded-arrays/ I think we need to change the schema to something like this:
```
{
    "user_id": "string",
    "review_id": "string",
    "timestamp": "string"
}
```
We need to update their queries in the application code. We also need to update their indexes and keys. The application logic is in ./src/UserReviewService/UserReviewHandler.h and ./src/MovieReviewService/UserReviewHandler.cpp files.

# Setting env

Update docker-compose file: prev -> new

      dns-media:                                dns-media:
        image: defreitas/dns-proxy-server           image: defreitas/dns-proxy-server
        volumes:                                    cpuset: "1"
                                                    volumes:

isolate cpu with:

    sudo nano /etc/default/grub
        GRUB_CMDLINE_LINUX="isolcpus=5,33"
        GRUB_CMDLINE_LINUX_DEFAULT="isolcpus=5,33"
    sudo update-grub
    sudo reboot
    check sudo cat /sys/devices/system/cpu/isolated

Disable smt:

    echo off | sudo tee /sys/devices/system/cpu/smt/control

# Tune the benchmark

SLO is 10x 99% tail latency

# ARM

change all yg397 to abdu1998a in docker-compose.yml
if you get an error:

    cd docker/openresty-thrift
    sudo docker build -t yg397/openresty-thrift:xenial -f xenial/Dockerfile .
    cd docker/thrift-microservice-deps/cpp
    sudo docker build -t yg397/thrift-microservice-deps:xenial .
    cd .
    sudo docker build -t yg397/media-microservices .
