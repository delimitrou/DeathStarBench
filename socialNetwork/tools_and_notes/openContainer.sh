#!/bin/bash
# Author: Luke Hobieka
# Description: opens container with bash
# Requires: docker compose is installed

docker exec -it $1 bash