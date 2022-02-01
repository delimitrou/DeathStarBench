#!/bin/bash
#####################################
### Build multiarch Docker images ###
#####################################

cd $(dirname $0)/..


EXEC="docker buildx"

USER="igorrudyk1"

TAG="latest"

# ENTER THE ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)
$EXEC create --name mybuilder --use

for i in frontend geo profile rate recommendation reserve search user
do
  IMAGE=hotel_reserv_${i}_single_node
  echo Processing image ${IMAGE}
  cd $ROOT_FOLDER
  $EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile . --platform linux/arm64,linux/amd64 --push
  cd $ROOT_FOLDER

  echo
done


cd - >/dev/null
