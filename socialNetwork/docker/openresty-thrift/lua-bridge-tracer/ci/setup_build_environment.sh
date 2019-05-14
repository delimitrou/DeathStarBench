#!/bin/bash

set -e
apt-get update 
apt-get install --no-install-recommends --no-install-suggests -y \
                ca-certificates \
                build-essential \
                software-properties-common \
                cmake \
                pkg-config \
                git \
                automake \
                autogen \
                autoconf \
                libtool \
                ssh \
                wget \
                curl \
                unzip \
                libreadline6-dev \
                libncurses5-dev \
                python python-setuptools python-pip

pip install gcovr
