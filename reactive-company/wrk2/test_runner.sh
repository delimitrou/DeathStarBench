#!/bin/bash


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > t1c1d15.csv

for i in {100..2000..100}
do
  ./wrk -t 1 -c 1 -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R $i | tail -n 1 >> t1c1d15.csv
  ./wrk -t 1 -c 1 -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R $i > t1c1d15.txt
done


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > t1d15R10.csv

for i in {5..100..5}
do
  ./wrk -t 1 -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R 10 | tail -n 1 >> t1d15R10.csv
  ./wrk -t 1 -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R 10 > t1d15R10.txt
done


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > d15R50.csv

for i in {1..8..1}
do
  ./wrk -t $i -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R 50 | tail -n 1 >> d15R50.csv
  ./wrk -t $i -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts_stream.lua http://localhost:8080/stream -R 50 > d15R50.txt
done


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > t1c1d15-gb.csv

for i in {100..2000..100}
do
  ./wrk -t 1 -c 1 -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R $i | tail -n 1 >> t1c1d15-gb.csv
  ./wrk -t 1 -c 1 -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R $i > t1c1d15-gb.txt
done


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > t1d15R10-gb.csv

for i in {5..100..5}
do
  ./wrk -t 1 -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R 10 | tail -n 1 >> t1d15R10-gb.csv
  ./wrk -t 1 -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R 10 > t1d15R10-gb.txt
done


echo "latency.min,latency.mean,latency.max,summary.errors.status,summary.errors.timeout,summary.errors.connect" > d15R50-gb.csv

for i in {1..8..1}
do
  ./wrk -t $i -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R 50 | tail -n 1 >> d15R50-gb.csv
  ./wrk -t $i -c $i -d 15 -L -s ./scripts/reactive-company/get_all_blogposts.lua http://localhost:8080/stream -R 50 > d15R50-gb.txt
done
