#!/bin/bash
# Author: Luke Hobieka
# Description: flushes all memcache instances in the deathstart benchmark suit
# Usage : flushMemcache.sh port1 port2 ... port n
# Requirements : netcat is installed
# TODO: take ports in as arguements

for var in "$@"
do
    echo 'flush_all' | nc localhost {$var}
done
