#!/bin/bash

# get command line arguments with flags
threads=1
connections=32
duration=2m
rps=2000 # icelake 2500, Zen3 2000, Ampere 950 should work but main idea is to fix the tail latency to ~100ms
save=0
ip=localhost

while getopts t:c:d:r:s:h:i: flag
do
    case "${flag}" in
        i) ip=${OPTARG};;
        t) threads=${OPTARG};;
        c) connections=${OPTARG};;
        d) duration=${OPTARG};;
        r) rps=${OPTARG};;
        s) save=${OPTARG};;
        # if flag is h print help message for how to use this script
        h) echo "Usage: $0 [-t <num-threads>] [-c <num-conns>] [-d <duration ex: 2m>] [-i <ip-addr>] [-r <reqs-per-sec>] [-s <save-results 1 or 0>] [-h]"; exit 1;;
        # if flag is unrecognized print help message for how to use this script
        \?) echo "Usage: $0 [-t <num-threads>] [-c <num-conns>] [-d <duration>] [-i <ip-addr>] [-r <reqs-per-sec>] [-s <save-results 1 or 0>] [-h]"; exit 1;;
    esac
done

# if save flag is 1 then send save result flags in script 
if [ $save -eq 1 ]
then
    ../wrk2/wrk -D exp -t $threads -c $connections -d $duration -L -s ./wrk2/scripts/media-microservices/compose-review-record.lua http://$ip:8080/wrk2-api/review/compose -R $rps

    curl http://$ip:9999/d
    curl http://$ip:9999/q
else
    ../wrk2/wrk -D exp -t $threads -c $connections -d $duration -L -s ./wrk2/scripts/media-microservices/compose-review.lua http://$ip:8080/wrk2-api/review/compose -R $rps
fi