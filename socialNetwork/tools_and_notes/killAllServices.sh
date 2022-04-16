#!/bin/bash
# Author: Luke Hobieka
# Description: kills all docker services
# Requires: docker compose is installed

docker kill $(docker ps -q)
