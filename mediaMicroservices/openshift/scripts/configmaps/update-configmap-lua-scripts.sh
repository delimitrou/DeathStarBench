#!/bin/bash

NS="media-microsvc"

cd $(dirname $0)/../..

oc create cm configmap-lua-scripts   --from-file=configmaps/lua-scripts --dry-run --save-config -o yaml -n ${NS} | oc apply -f - -n ${NS}

for i in cast-info movie movie-info plot review user
do
  oc create cm configmap-lua-scripts-${i} --from-file=configmaps/lua-scripts/wrk2-api/${i} --dry-run --save-config -o yaml -n ${NS} | oc apply -f - -n ${NS}
done

cd -
