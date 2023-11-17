#!/bin/bash

docker-compose down
docker volume rm $(docker volume ls -q)
docker-compose up -d

python3 scripts/write_movie_info.py -c datasets/tmdb/casts.json -m datasets/tmdb/movies.json --server_address http://localhost:8080 && scripts/register_users.sh && scripts/register_movies.sh