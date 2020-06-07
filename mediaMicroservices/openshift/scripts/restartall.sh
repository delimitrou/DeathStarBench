#!/bin/bash

NS=media-microsvc

work="cast-info-memcached cast-info-mongodb cast-info-service compose-review-memcached compose-review-service jaeger media-microsvc-ns movie-id-memcached movie-id-mongodb movie-id-service movie-info-memcached movie-info-mongodb movie-info-service movie-review-mongodb movie-review-redis movie-review-service nginx-web-server plot-memcached plot-mongodb plot-service rating-redis rating-service review-storage-memcached review-storage-mongodb review-storage-service text-service unique-id-service user-memcached user-mongodb user-review-mongodb user-review-redis user-review-service user-service mms-client"


function finish {
cd $(dirname $0)/../..

echo "Run the following to load the dataset"
echo
echo "cd /root/DeathStarBench/mediaMicroservices/scripts"
echo "python3 write_movie_info.py"
echo "./register_users.sh"
echo "cd -"

cd -
}

function usage()
{
    echo    "Script to restart containers and services"
    echo    ""
    echo -e "\t-h --help"
    echo -e "\t-s --show-update watch the pods creation (default true)"
    echo -e "\t-i --init-dataset load the dataset (default true)"
    echo -e "\tPlease, include space between the parameter and value and \"\" for multiple values"
    echo    ""
}

if [[ -z $@ ]]; then
  echo "No parameter specified, use '"-h"' to see all available configurations"
fi

while [ "$1" != "" ]; do
    PARAM=$1; shift
    VALUE=$1; shift
    case $PARAM in
        -h | --help)
            usage
            exit
            ;;
        -s | --show-update)
            SHOW_UPDATE=$VALUE
            ;;
        -i | --init-dataset)
            INIT_DATASET=$VALUE
            ;;
        *)
            echo "ERROR: unknown parameter \"$PARAM\""
            usage
            exit 1
            ;;
    esac
done

if [[ -z $INIT_DATASET ]] || [[ $(echo $INIT_DATASET | egrep "1|true" |wc -l)  -gt 0 ]]; then
  trap finish EXIT
fi

echo this may take a while ... use control-c when status screen shows all services up.
echo reduce replicas to 0
for d in ${work}
do
	oc scale --replicas=0 deployment/$d -n ${NS} &
done

wait

echo increase replicas back to 1
for d in ${work} 
do
	oc scale --replicas=1 deployment/$d -n ${NS} &
done

wait

if [[ -z $SHOW_UPDATE ]] || [[ $(echo $SHOW_UPDATE | egrep "1|true" |wc -l) -gt 0 ]]; then
  echo now wait for everything to come back up
  watch oc get pods -n ${NS}
else
  running=$(oc get pods -n ${NS} --no-headers | grep -c Running)
  total=$(echo ${work} | wc -w)
  while [[ $running -lt ${total} ]]; do
	  echo "Waiting for $((total-running)) more pods to start"
    sleep 1
    running=$(oc get pods -n ${NS} --no-headers | grep -c Running)
  done
fi
