#!/bin/bash

cd $(dirname $0)/..


EXEC=docker

USER="salehsedghpour"

TAG="latest"

# ENTER THE ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)

for i in frontend geo profile rate recommendation reserve search user
do
  IMAGE=hotel_reserv_${i}_single_node
  echo Processing image ${IMAGE}
  cd $ROOT_FOLDER
  $EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile .
  $EXEC push "$USER"/"$IMAGE":"$TAG"
  cd $ROOT_FOLDER

  echo
done


cd - >/dev/null
