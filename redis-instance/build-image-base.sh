#!/bin/bash

SH_FILES=`find . -name "*.sh"`
for FILE in ${SH_FILES[*]}
do
    dos2unix $FILE
    chmod +x $FILE
done

HUB=harbor.ceclouddyn.com/paas
IMAGE_NAME=operator-redis-ha-instance-base
TAG=0.0.26

# nice -n -20 docker buildx build --no-cache --platform linux/amd64,linux/arm64 -f ./dockerfile-base \
#     -t $HUB/$IMAGE_NAME:$TAG --push . 

nice -n -20 docker buildx build --platform linux/amd64 -f ./dockerfile-base \
    -t $HUB/$IMAGE_NAME:$TAG --push . 
