#!/bin/bash

cd $(dirname $0)/../..

oc create cm configmap-nginx-conf   --from-file=configmaps/nginx.conf  -n media-microsvc

cd -
