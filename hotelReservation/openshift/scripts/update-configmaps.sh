#!/bin/bash

cd $(dirname $0)/..

NS="hotel-res"

oc create cm configmap-config-json   --from-file=configmaps/config.json  --dry-run --save-config -o yaml -n ${NS} | oc apply -f - -n ${NS}

cd - >/dev/null

