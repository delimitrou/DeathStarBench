#!/bin/bash

NS=social-network

cd $(dirname $0)/../..

# This script creates an OpenShift ConfigMap for all the services
# built upon the C++ jaeger client, which uses the jaeger-config.yml
# to find the jaeger end-point URL.
oc create cm jaeger-config-yaml   --from-file=config/jaeger-config.yml               -n ${NS} --dry-run --save-config -o yaml | oc apply -f - -n ${NS}

# Since the nginx-thrift service is not built upon the C++ jaeger client,
# this service requires the jaeger-config.json in a different format than
# the one in the ConfigMap jaeger-config. Then, we create a new ConfigMap.
oc create cm nginx-thrift-jaeger --from-file=nginx-thrift-config/jaeger-config.json -n ${NS} --dry-run --save-config -o yaml | oc apply -f - -n ${NS}
