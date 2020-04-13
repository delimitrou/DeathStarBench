#!/bin/bash

NS=hotel-res

cd $(dirname $0)/..

oc create namespace ${NS} 2>/dev/null
oc project ${NS}

oc adm policy add-scc-to-user anyuid -z default -n ${NS}
oc adm policy add-scc-to-user privileged -z default -n ${NS}
oc policy add-role-to-user system:image-puller system:serviceaccount:hotel-res:default -n hotel-res
oc policy add-role-to-user system:image-puller kube:admin -n hotel-res

oc policy add-role-to-user system:image-builder kube:admin -n hotel-res
oc policy add-role-to-user registry-viewer kube:admin -n hotel-res
oc policy add-role-to-user registry-editor kube:admin -n hotel-res

./scripts/create-configmaps.sh
for i in *.yaml
do
  oc apply -f ${i} -n ${NS} &
done
wait

echo "Finishing in 30 seconds"
sleep 30

oc get pods -n ${NS}

cd - >/dev/null

