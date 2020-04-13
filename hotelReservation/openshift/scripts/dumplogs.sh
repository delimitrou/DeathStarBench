#!/bin/bash

NS="hotel-res"

work="consul frontend geo jaeger memcached-profile memcached-rate memcached-reserve mongodb-geo mongodb-profile mongodb-rate mongodb-recommendation mongodb-reservation mongodb-user profile rate recommendation reservation search user"

for d in ${work}
do
	oc logs deployment/${d} --all-containers -n ${NS} > ${d}.log
done


wait
