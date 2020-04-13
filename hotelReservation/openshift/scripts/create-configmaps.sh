#!/bin/bash

cd $(dirname $0)/..

NS=hotel-res

oc create cm configmap-config-json   --from-file=configmaps/config.json  -n ${NS}

cd - >/dev/null
