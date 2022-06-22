#!/bin/bash

set -e

./sbt microservice_1/docker:publishLocal
./sbt microservice_2/docker:publishLocal
#docker-compose up
