#!/bin/bash

cd $(dirname $0)/..

./scripts/configmaps/update-configmap-gen-lua.sh
./scripts/configmaps/update-configmap-jaeger-config-json.sh
./scripts/configmaps/update-configmap-lua-scripts.sh
./scripts/configmaps/update-configmap-nginx-conf.sh

cd -
