#!/bin/bash

cd $(dirname $0)/..

./scripts/configmaps/create-jaeger-configmap.sh
./scripts/configmaps/create-media-frontend-configmap.sh
./scripts/configmaps/create-nginx-thrift-configmap.sh

cd - >/dev/null
