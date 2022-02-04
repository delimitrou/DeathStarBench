#!/bin/bash

cd $(dirname $0)/..


EXEC=docker

USER="supreethkurpad"

TAG="latest"

# ENTER THE ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)

for i in hotelreservation #frontend geo profile rate recommendation reserve search user #uncomment to build multiple images
do
  IMAGE=${i}
  echo Processing image ${IMAGE}
  cd $ROOT_FOLDER
  $EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile .
  $EXEC push "$USER"/"$IMAGE":"$TAG"
  cd $ROOT_FOLDER

  echo
done


cd - >/dev/null
