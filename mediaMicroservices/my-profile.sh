#!/bin/bash

CORES=($(seq 12 1 16))
NAMES=("compose-review-service" "compose-review-memcached" "review-storage-service" "review-storage-mongodb" "review-storage-memcached")
# CORES+=(32 33)
# NAMES+=("movie-info-service" "movie-info-mongodb")

COUNTERS=("cycles:u,cycles:k,instructions:u,instructions:k")
METRICS=("DSB_Coverage")
TOPDOWN=1

echo ${CORES[@]}
echo ${NAMES[@]}
echo "init: $1"
exit

PLATFORM="icelake"
OUT="OUT"
iterations=50
TOPDOWNiter=5
DEV=1
INIT=$1
RESULTS="results"

PERIOD=60
CORESlength=${#CORES[@]}

function log_folder () {
    if [[ ! -d $RESULTS ]]; then
        (($DEV)) && echo "create experimental folder $RESULTS"
        mkdir $RESULTS
    fi

    if [[ ! -d ${RESULTS}/${PLATFORM} ]]; then
        (($DEV)) && echo "create experimental folder ${RESULTS}/${PLATFORM}"
        mkdir ${RESULTS}/${PLATFORM}
    fi


    if [[ ! -d $OUT ]]; then
        (($DEV)) && echo "create tmp folder $OUT"
        mkdir $OUT
    else
        exp_cnt=`ls ${RESULTS}/${PLATFORM} | grep -Eo [0-9]+ | sort -rn | head -n 1`
        (($DEV)) && echo "max exp count is $exp_cnt"
        [ "$(ls -A $OUT)" ] && mv $OUT $RESULTS/$PLATFORM/$((exp_cnt + 1)) && mkdir $OUT
    fi
}

if (( INIT == 1 )); then
    python3 scripts/write_movie_info.py -c datasets/tmdb/casts.json -m datasets/tmdb/movies.json --server_address http://localhost:8080 && scripts/register_users.sh && scripts/register_movies.sh

    exit
fi
function do_perf(){
    log_folder
    for (( i = 0; i < iterations; i++ ))
    do
        echo "iteration $i"
        for (( j=0; j < CORESlength; j++ ))
        do
            PERF_FILE=${OUT}/${NAMES[j]}/"perf.log"
            TOPDOWN_FILE=${OUT}/${NAMES[j]}/"topdown.log"
            echo "core $j file: $PERF_FILE"
            if (( TOPDOWN == 1 && i < TOPDOWNiter)); then
                /var/services/homes/shanqing-epfl/pmu-tools/toplev -l3 --core C${CORES[j]} -v --no-desc --no-multiplex -- sleep ${PERIOD} >> ${TOPDOWN_FILE} 2>&1
            fi
            for k in ${COUNTERS[@]}; 
            do
                perf stat -C ${CORES[j]} -e $k --output ${PERF_FILE} --append -- sleep ${PERIOD} 
            done
            for k in ${METRICS[@]}; 
            do
                perf stat -C ${CORES[j]} -M $k --output ${PERF_FILE} --append -- sleep ${PERIOD} 
            done
        done
        sleep 1
    done
}

do_perf 