#!/bin/bash

cd $(dirname $0)/../..

NS="media-microsvc"

oc create cm configmap-jaeger-config-json   --from-file=configmaps/jaeger-config.json --dry-run --save-config -o yaml -n ${NS} | oc apply -f - -n ${NS}

cd -
