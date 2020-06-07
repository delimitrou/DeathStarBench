#!/bin/bash

# tip to install ts command:
# mac: brew install moreutils --without-parallel
#      brew install parallel
# ubuntu: apt-get install moreutils
# rhel: yum install moreutils

trap 'kill $(jobs -pr) >/dev/null 2>&1' SIGINT SIGTERM EXIT

command -v ts >/dev/null
NOTS=${?}

for d in compose-post-redis compose-post-service home-timeline-redis home-timeline-service media-frontend media-memcached media-mongodb media-service nginx-thrift post-storage-memcached post-storage-mongodb post-storage-service social-graph-mongodb social-graph-redis social-graph-service text-service unique-id-service url-shorten-memcached url-shorten-mongodb url-shorten-service user-memcached user-mention-service user-mongodb user-service user-timeline-mongodb user-timeline-redis user-timeline-service write-home-timeline-rabbitmq write-home-timeline-service
do
	if [[ ${NOTS} -eq 1 ]]
	then
		oc logs -f deployment/${d} --all-containers -n social-network &
	else
		oc logs -f deployment/${d} --all-containers -n social-network | ts "${d}: " &
	fi
done


wait
