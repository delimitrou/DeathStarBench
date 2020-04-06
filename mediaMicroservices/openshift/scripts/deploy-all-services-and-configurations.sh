#!/bin/bash

cd $(dirname $0)/..
NS="media-microsvc"

oc create namespace ${NS}
oc project ${NS}

oc adm policy add-scc-to-user anyuid -z default -n ${NS}
oc adm policy add-scc-to-user privileged -z default -n ${NS}

./scripts/create-all-configmap.sh 
./scripts/create-destination-rule-all.sh

for service in *.yaml
do
  oc apply -f $service -n ${NS}
done

echo "Finishing in 30 seconds"
sleep 30

mmsclient=$(oc get pod | grep mms-client- | cut -f 1 -d " ")

echo "After all pods have been created:"
echo
echo "Verify that files under DeathStarBench/mediaMicroservices/scripts have the latest web server url -- or use the local cluster addressing scheme: nginx-web-server.media-microsvc.svc.cluster.local"
echo "oc get ep | grep nginx-web-server"
oc get ep | grep nginx-web-server
echo "Files containing known http endpoints:"
grep "http:" ../scripts/* 
echo
echo "Make mms-client useful with the following command:"
echo "oc cp /root/DeathStarBench media-microsvc/"${mmsclient}":/root"
echo
echo "You can log into the mms-client with this command:"
echo "oc rsh deployment/mms-client"
echo
echo "Run the following on a working mms client node to load the dataset:"
echo
echo "cd /root/DeathStarBench/mediaMicroservices/scripts"
echo "python3 write_movie_info.py"
echo "./register_users.sh"
echo
echo "cd /root/DeathStarBench/mediaMicroservices/wrk2"
echo "make clean"
echo "make"
echo
echo "Finally, load the system with the wrk2 command, from the previous directory, using the correct web server address..."
echo "./wrk -D exp -t 2 -c 4 -d 32 -L -s ./scripts/media-microservices/compose-review.lua http://nginx-web-server.media-microsvc.svc.cluster.local:8080/wrk2-api/review/compose -R 8"


cd - >/dev/null
