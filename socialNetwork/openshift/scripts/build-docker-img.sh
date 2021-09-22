#!/bin/bash
# ##### TROUBLESHOOTING ERRORS ######
# Make sure you are connected to your OpenShift docker REGISTRY. You can verify which node it is running with the following command:
# oc get pods -n default -o wide | grep -v console | grep registry
# And then, you can add it to your /etc/hosts accordingly
#
# If you are getting the error: registry could not be contacted at default-route-openshift-image-registry.apps.xyz.com: Get https://default-route-openshift-image-registry.apps.xyz.com/v2/: x509: certificate signed by unknown authority
# One way to fix this problem is setting insecure registry config to docker daemon ~/.docker/daemon.json, after that, please make sure to restart docker daemon. Please change your REGISTRY url accordingly.
# {
#  "debug" : true,
#  "experimental" : true,
#  "insecure-registries" : [
#     "default-route-openshift-image-registry.apps.xyz.com"
#  ]
# }
# or if you are using podman edit/create the /etc/containers/registries.conf, please change your REGISTRY url accordingly
# [registries.insecure]
# registries = ['docker-registry-default.apps.ocp-x86-2.mycluster']
# [registries.search]
# registries = ['docker.io', 'registry.fedoraproject.org', 'registry.access.redhat.com']
# and/or add the use image-builder group
# oc policy add-role-to-user system:image-builder $(oc whoami)
# oc policy add-role-to-user edit $(oc whoami)
# You may also need to add a role in the cluster admin group
# oc adm policy add-cluster-role-to-user cluster-admin $(oc whoami)
#
# If you see a 500 error, if might be that the social-network namespace is not created, then create it with the following command:
# oc create namespace social-network

#####
EXEC=docker
if [[ $(podman -v | grep version | wc -l) -gt 0 ]]; then
  EXEC=podman
else
  echo "Using docker, but we recommend to use podman"
fi


TAG="openshift"
PROJECT="social-network"

cd $(dirname $0)/..

# LOGIN IN THE $EXEC REGISTRY FOR OPENSHIFT
REGISTRY=$(oc get route -n default | grep registry | grep -v console | awk '{print $2}')
TOKEN=$(oc whoami -t)
USER=$(oc whoami)
oc project $PROJECT
oc registry login \
  --insecure=true --skip-check -z default --token=$TOKEN $REGISTRY
$EXEC login -u $USER -p $TOKEN $REGISTRY

# ENTER IN THE SOCIAL-NETWORK ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)

# BUILD MEDIA FRONTEND IMAGE
SERVICE="media-frontend"
if [[ $($EXEC images | grep $SERVICE | wc -l) -le 0 ]]; then
  cd docker/media-frontend/
  $EXEC build -t "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG" -f xenial/Dockerfile .
  cd $ROOT_FOLDER
else
  echo "$SERVICE image already exist"
fi
$EXEC push "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG"

# BUILD OPENRESTY-THRIFT
SERVICE="openresty-thrift"
if [[ $($EXEC images | grep $SERVICE | wc -l) -le 0 ]]; then
  cd docker/openresty-thrift/
  $EXEC build -t "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG" -f xenial/Dockerfile .
  cd $ROOT_FOLDER
else
  echo "$SERVICE image already exist"
fi
$EXEC push "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG"

# BUILD SOCIAL-NETWORK MICROSERVICE DEPS
SERVICE="thrift-microservice-deps"
if [[ $($EXEC images | grep $SERVICE | wc -l) -le 0 ]]; then
  cd docker/thrift-microservice-deps
  $EXEC build -t "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG" -f cpp/Dockerfile .
else
  echo "$SERVICE image already exist"
fi

# BUILD SOCIAL-NETWORK MICROSERVICE
SERVICE="social-network-microservices"
if [[ $($EXEC images | grep $SERVICE | wc -l) -le 0 ]]; then
  cd $ROOT_FOLDER
  $EXEC build -t "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG" .
else
  echo "$SERVICE image already exist"
fi
$EXEC push "$REGISTRY"/"$PROJECT"/"$SERVICE":"$TAG"

echo "Images:"
$EXEC images | grep "$TAG"
