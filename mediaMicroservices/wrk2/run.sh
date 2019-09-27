#!/bin/bash

## A script to run a test

RPS=$1
IP=128.253.128.76
PORT=88
PN=/home/sc2682/client/wrk2
MEASURE=20
CONN=400

nice -n -20 ./wrk -c $CONN -t 40 -d $MEASURE -s scripts/multiplepaths.lua --timeout 10000000 -R $RPS http://$IP:$PORT


