#!/bin/bash

set -e

# Docker Machine Setup - master
docker-machine create \
 	-d virtualbox \
 	swmaster
# Docker Machine Setup - node 1
docker-machine create \
 	-d virtualbox \
 	swnode1
# Docker Machine Setup - node 2
docker-machine create \
 	-d virtualbox \
 	swnode2
# Docker Machine Setup - node 3
 docker-machine create \
 	-d virtualbox \
 	swnode3


# Configure swarm mode cluster - initialization on master
eval $(docker-machine env swmaster)
docker swarm init --advertise-addr $(docker-machine ip swmaster):2377
jointoken=$(docker swarm join-token --quiet worker)

# Configure swarm mode cluster - join nodes
eval $(docker-machine env swnode1)
docker swarm join --token $jointoken $(docker-machine ip swmaster):2377
eval $(docker-machine env swnode2)
docker swarm join --token $jointoken $(docker-machine ip swmaster):2377
eval $(docker-machine env swnode3)
docker swarm join --token $jointoken $(docker-machine ip swmaster):2377



#List all nodes
eval $(docker-machine env swmaster)
echo "-------------------------"
echo "######### Nodes: ########"
echo "-------------------------"
docker node ls

# Create a stack using docker deploy command
docker stack deploy --compose-file docker-compose.yml stack

# List all services
eval $(docker-machine env swmaster)
echo "-------------------------"
echo "####### Services: #######"
echo "-------------------------"
docker service ls

# Explore the API
echo "-------------------------"
echo "#########################"
echo "#### Explore the API: ###"
echo "#########################"
echo "Please have patience, containers need some time to start. You can run this command to list all your services (with current status): docker service ls"
echo "-------------------------"
echo "-------------------------------------------------------------"
echo "-------------------------"
echo "### Create Blog Post: ###"
echo "-------------------------"
echo "curl -H \"Content-Type: application/json\" -X POST -d '{\"title\":\"xyz\",\"rawContent\":\"xyz\",\"publicSlug\": \"publicslug\",\"draft\": true,\"broadcast\": true,\"category\": \"ENGINEERING\", \"publishAt\": \"2016-12-23T14:30:00+00:00\"}' $(docker-machine ip swmaster):$(docker service inspect --format='{{ (index (index .Endpoint.Ports) 0).PublishedPort}}'  stack_reactive-company)/blogposts"
echo "-----------------------------"
echo "#### Read all blog posts: ####"
echo "-----------------------------"
echo "curl $(docker-machine ip swmaster):$(docker service inspect --format='{{ (index (index .Endpoint.Ports) 0).PublishedPort}}'  stack_reactive-company)/blogposts"

echo "-------------------------------"
echo "######## Have fun ! ###########"
echo "-------------------------------"
echo "### Scale service  'scale stack_reactive-company' ###"
echo "eval \$(docker-machine env swmaster)"
echo "docker service scale stack_reactive-company=2"
echo "docker service ps stack_reactive-company"
