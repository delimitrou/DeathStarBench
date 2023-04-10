# Initial setup
sudo apt install python3-pip

pip3 install aiohttp, asyncio

sudo apt-get install libssl-dev

sudo apt-get install libz-dev

sudo apt-get install luarocks

sudo luarocks install luasocket

sudo apt  install docker-compose

cd wrk2

make all

cd ../mediaMicroservices

sudo docker-compose up -d

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
cd docker/openresty-thrift
sudo docker build -t yg397/openresty-thrift:xenial -f xenial/Dockerfile .
docker/thrift-microservice-deps/cpp
sudo docker build -t yg397/thrift-microservice-deps:xenial .
