#!/bin/bash

cd $(dirname $0)/..

./scripts/configmaps/create-configmap-jaeger-config-json.sh
./scripts/configmaps/create-configmap-lua-scripts.sh
./scripts/configmaps/create-configmap-nginx-conf.sh
./scripts/configmaps/create-configmap-gen-lua.sh

cd -
