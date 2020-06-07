#!/bin/bash

# tip to install ts command:
# mac: brew install moreutils --without-parallel
#      brew install parallel
# ubuntu: apt-get install moreutils
# rhel: yum install moreutils

trap 'kill $(jobs -pr) >/dev/null 2>&1' SIGINT SIGTERM EXIT

NS=media-microsvc

work="cast-info-memcached cast-info-mongodb cast-info-service compose-review-memcached compose-review-service jaeger media-microsvc-ns movie-id-memcached movie-id-mongodb movie-id-service movie-info-memcached movie-info-mongodb movie-info-service movie-review-mongodb movie-review-redis movie-review-service nginx-web-server plot-memcached plot-mongodb plot-service rating-redis rating-service review-storage-memcached review-storage-mongodb review-storage-service text-service unique-id-service user-memcached user-mongodb user-review-mongodb user-review-redis user-review-service user-service mms-client"


command -v ts >/dev/null
NOTS=${?}

for d in ${work}
do
	if [[ ${NOTS} -eq 1 ]]
	then
		oc logs -f deployment/${d} --all-containers -n ${NS} &
	else
		oc logs -f deployment/${d} --all-containers -n ${NS} | ts "${d}: " &
	fi
done


wait
