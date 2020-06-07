#!/bin/bash

cd $(dirname $0)/..

./scripts/configmaps/update-jaeger-configmap.sh
./scripts/configmaps/update-media-frontend-configmap.sh
./scripts/configmaps/update-nginx-thrift-configmap.sh

cd - >/dev/null
