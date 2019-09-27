#!/usr/bin/env bash

declare -a services=( \
    "user" \
    "movie-id" \
    "review-storage" \
    "user-review" \
    "movie-review" \
)

num_config=1
num_router=1
num_shard=3
num_replica=2

for service in $services; do
  if [ $num_config = 1 ]; then
    docker-compose exec $service-mongodb-config bash -c "mongo < /scripts/init-config.js"
  else
    for i in $(seq $num_config); do
      docker-compose exec $service-mongodb-config_$i bash -c "mongo < /scripts/init-config.js"
    done
  fi

  for i in $(seq $num_shard); do
    if [ $num_replica = 1 ]; then
      docker-compose exec $service-mongodb-shard-$i bash -c "mongo < /scripts/init-shard_$i.js"
    else
      for j in $(seq $num_replica); do
        docker-compose exec $service-mongodb-shard-$i\_$j bash -c "mongo < /scripts/init-shard_$i.js"
      done
    fi
  done

  sleep 10

  if [ $num_router = 1 ]; then
    docker-compose exec $service-mongodb-router bash -c "mongo < /scripts/init-router.js"
  else
    for i in $(seq $num_router); do
      docker-compose exec $service-mongodb-router_$i bash -c "mongo < /scripts/init-router.js"
    done
  fi
done

