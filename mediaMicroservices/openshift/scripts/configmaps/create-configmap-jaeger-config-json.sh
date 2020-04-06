#!/bin/bash

cd $(dirname $0)/../..

oc create cm configmap-jaeger-config-json   --from-file=configmaps/jaeger-config.json -n media-microsvc

cd -
