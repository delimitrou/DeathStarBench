#!/bin/bash

NS=social-network

echo "Are you really sure you want to erase the ${NS} world? (y/n)"
read a

if [[ "${a}z" != "yz" ]]
then
	echo "The population of the world thanks you!"
	exit 1
fi

echo "Are you really really really sure? (y/n)"
read a

if [[ "${a}z" != "yz" ]]
then
	echo "Doomsday narrowly averted..."
	exit 1
fi

work="compose-post-redis compose-post-service home-timeline-redis home-timeline-service media-frontend media-memcached media-mongodb media-service nginx-thrift post-storage-memcached post-storage-mongodb post-storage-service social-graph-mongodb social-graph-redis social-graph-service text-service unique-id-service url-shorten-memcached url-shorten-mongodb url-shorten-service user-memcached user-mention-service user-mongodb user-service user-timeline-mongodb user-timeline-redis user-timeline-service write-home-timeline-rabbitmq write-home-timeline-service"


echo deleting services and deployments

oc project ${NS}

for d in ${work}
do
	oc delete service/$d -n ${NS} &
 	oc delete deployment/$d -n ${NS} &
#	oc delete pod/$d -n ${NS} &
done

wait

echo deleting cm
for c in jaeger-config-yaml media-frontend-lua media-frontend-nginx nginx-thrift-jaeger nginx-thrift-genlua nginx-thrift-pages nginx-thrift-luascripts nginx-thrift-luascripts-api-home-timeline nginx-thrift-luascripts-api-post nginx-thrift-luascripts-api-user nginx-thrift-luascripts-api-user-timeline nginx-thrift-luascripts-wrk2-api-home-timeline nginx-thrift-luascripts-wrk2-api-post nginx-thrift-luascripts-wrk2-api-user nginx-thrift-luascripts-wrk2-api-user-timeline nginx-thrift

do
	oc delete cm/${c} -n ${NS}
done

echo finally deleting namespace ${NS}
oc delete namespace/${NS}

echo run deploy-all-services-and-configurations.sh to deploy again.
