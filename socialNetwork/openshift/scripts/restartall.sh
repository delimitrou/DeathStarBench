#!/bin/bash

NS="social-network"

function finish {
cd $(dirname $0)/../..

echo initialize the dataset
python3 scripts/init_social_graph.py

cd -
}

function usage()
{
    echo    "Script to restart the services"
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
for d in compose-post-redis compose-post-service home-timeline-redis home-timeline-service media-frontend media-memcached media-mongodb media-service nginx-thrift post-storage-memcached post-storage-mongodb post-storage-service social-graph-mongodb social-graph-redis social-graph-service text-service unique-id-service url-shorten-memcached url-shorten-mongodb url-shorten-service user-memcached user-mention-service user-mongodb user-service user-timeline-mongodb user-timeline-redis user-timeline-service write-home-timeline-rabbitmq write-home-timeline-service
do
	oc scale --replicas=0 deployment/$d -n ${NS} &
done

wait

echo increase replicas back to 1
for d in compose-post-redis compose-post-service home-timeline-redis home-timeline-service media-frontend media-memcached media-mongodb media-service nginx-thrift post-storage-memcached post-storage-mongodb post-storage-service social-graph-mongodb social-graph-redis social-graph-service text-service unique-id-service url-shorten-memcached url-shorten-mongodb url-shorten-service user-memcached user-mention-service user-mongodb user-service user-timeline-mongodb user-timeline-redis user-timeline-service write-home-timeline-rabbitmq write-home-timeline-service
do
	oc scale --replicas=1 deployment/$d -n ${NS} &
done

wait

if [[ -z $SHOW_UPDATE ]] || [[ $(echo $SHOW_UPDATE | egrep "1|true" |wc -l) -gt 0 ]]; then
  echo now wait for everything to come back up
  watch oc get pods -n ${NS}
else
  running=$(oc get pods -n ${NS} --no-headers | grep Running |wc -l)
  while [[ $running -lt 30 ]]; do
    echo "Waiting $(echo 30-$running|bc) pods to start"
    sleep 1
    running=$(oc get pods -n ${NS} --no-headers | grep Running |wc -l)
  done
fi
