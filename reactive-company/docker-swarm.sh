#!/bin/bash

set -e

## Enable swarm mode.
## Other nodes can join this swarm cluster and this would easily allow to deploy the multi-container application to a multi-host as well.
docker swarm init
## Deploy the services defined in Compose file.
docker stack deploy --compose-file=docker-compose.yml stack


