#!/bin/bash

# ##### DEBUGGING ERROR ######
# If you are getting the error: registry could not be contacted at default-route-openshift-image-registry.apps.xyz.com: Get https://default-route-openshift-image-registry.apps.xyz.com/v2/: x509: certificate signed by unknown authority
# One way to fix this problem is setting insecure registry config to docker daemon ~/.docker/daemon.json, after that, please make sure to restart docker daemon.
# Replace abc.xyz with local value.
#{
#  "debug" : true,
#  "experimental" : true,
#  "insecure-registries" : [
#     "default-route-openshift-image-registry.apps.xyz.com"
#  ]
#}
# #####

# When rerunning this script first remove existing images:
# e.g.,
# podman images
# podman image rm -f xyz
# and
# oc get images | grep hotel-res
# oc delete image/sha256:xyz


cd $(dirname $0)/..

NS=hotel-res

EXEC=docker
TLSVERIFY=""
command -v podman >/dev/null
NOPODMAN=${?}

if [[ ${NOPODMAN} -eq  0 ]]; then
  EXEC=podman
  TLSVERIFY="--tls-verify=false"
else
  echo "Using docker, but we recommend to use podman"
fi


#TAG="openshift"
TAG="latest"
PROJECT="hotel-res"

# LOGIN IN THE $EXEC REGISTRY FOR OPENSHIFT
TOKEN=$(oc whoami -t)
oc project $PROJECT
oc registry login \
  --insecure=true --skip-check -z default --token=$TOKEN \
  image-registry-openshift-image-registry.apps.cogadvisor.openshiftv4test.com
$EXEC login --namespace ${NS} ${TLSVERIFY} \
  -p $TOKEN -u kubeadmin $(oc registry info)

REGISTRY=$(oc registry info)

# ENTER THE ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)

for i in frontend geo profile rate recommend rsv search user
do
  IMAGE=hotel_reserv_${i}_single_node
  echo Processing image ${IMAGE}
  if [[ $($EXEC images --namespace ${NS} | grep $IMAGE | wc -l) -le 0 ]]; then
    oc create imagestream ${IMAGE} -n ${NS} 2>/dev/null
    cd $ROOT_FOLDER
    $EXEC build --namespace ${NS} ${TLSVERIFY} -t "$REGISTRY"/"$PROJECT"/"$IMAGE":"$TAG" -f Dockerfile .
    $EXEC push --namespace ${NS} ${TLSVERIFY} "$REGISTRY"/"$PROJECT"/"$IMAGE":"$TAG"
    cd $ROOT_FOLDER
  else
    echo "$IMAGE image already exists"
  fi
  echo
done

echo "Images:"
$EXEC images --namespace ${NS} | grep "$TAG"

cd - >/dev/null
