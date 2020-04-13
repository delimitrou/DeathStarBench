#!/bin/bash

work="cast-info-memcached cast-info-mongodb cast-info-service compose-review-memcached compose-review-service jaeger media-microsvc-ns movie-id-memcached movie-id-mongodb movie-id-service movie-info-memcached movie-info-mongodb movie-info-service movie-review-mongodb movie-review-redis movie-review-service nginx-web-server plot-memcached plot-mongodb plot-service rating-redis rating-service review-storage-memcached review-storage-mongodb review-storage-service text-service unique-id-service user-memcached user-mongodb user-review-mongodb user-review-redis user-review-service user-service mms-client"

for d in ${work}
do
	oc logs deployment/${d} --all-containers -n media-microsvc > ${d}.log
done


wait
