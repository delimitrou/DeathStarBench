#!/bin/bash

cd $(dirname $0)/..

NS="media-microsvc"

FOLDER="networking"
FILE="destination-rule-all.yaml"

if [[ ! -f ${FOLDER}/${FILE} ]]; then
  echo "The file ${FOLDER}/${FILE} does not exist"
  echo "You can use the script scripts/helper/generate-destination-rule-all-services.sh to create it"
fi

oc apply -f ${FOLDER}/${FILE} -n ${NS}

cd -
