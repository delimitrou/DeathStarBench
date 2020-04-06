#!/bin/bash

cd $(dirname $0)/../..

NS="media-microsvc"

oc create cm configmap-gen-lua   --from-file=configmaps/gen-lua --dry-run --save-config -o yaml -n ${NS} | oc apply -f - -n ${NS}

cd -
