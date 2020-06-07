#!/bin/bash

for d in compose-post-redis compose-post-service home-timeline-redis home-timeline-service media-frontend media-memcached media-mongodb media-service nginx-thrift post-storage-memcached post-storage-mongodb post-storage-service social-graph-mongodb social-graph-redis social-graph-service text-service unique-id-service url-shorten-memcached url-shorten-mongodb url-shorten-service user-memcached user-mention-service user-mongodb user-service user-timeline-mongodb user-timeline-redis user-timeline-service write-home-timeline-rabbitmq write-home-timeline-service
do
	oc logs deployment/${d} --all-containers -n social-network > ${d}.log
done


wait
