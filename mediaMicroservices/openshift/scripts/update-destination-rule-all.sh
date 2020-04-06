#!/bin/bash

cd $(dirname $0)/..

NS="media-microsvc"
FOLDER="networking"
FILE="destination-rule-all.yaml"

if [[ ! -f ${FOLDER}/${FILE} ]]; then
  echo "The file $FILE does not exist"
  echo "You can use the script scripts/helper/generate-destination-rule-all-services.sh to create it"
fi

oc create -f ${FOLDER}/${FILE} --dry-run --save-config -o yaml --namespace ${NS} | oc apply -f - --namespace ${NS}

cd -

