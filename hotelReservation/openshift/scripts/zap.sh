#!/bin/bash

NS=hotel-res

cd $(dirname $0)/..


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


cd ..

for s in consul frontend geo jaeger memcached-profile memcached-rate memcached-reserve mongodb-geo mongodb-profile mongodb-rate mongodb-recommendation mongodb-reservation mongodb-user profile rate recommendation reservation search user
do
 	oc delete service/$s -n ${NS} &
 	oc delete deployment/$s -n ${NS} &
done
oc delete deployment/hr-client -n ${NS} &
wait

#	oc delete pod/$d -n ${NS} &

for i in geo profile rate recommendation reservation user
do
	oc delete pv/$i -n ${NS} &
done
wait

for c in configmap-config-json
do
        oc delete cm/${c} -n ${NS}
done

# echo finally deleting namespace ${NS}
# oc delete namespace/${NS}


cd - >/dev/null
echo Finishing in 5 seconds...
sleep 5

