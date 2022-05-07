#!/bin/bash

rm ./csiplugin -rf

ARCH=`uname -m`

echo "ARCH=$ARCH"

if [ "x86_64" == "$ARCH" ]; then
    echo "cp ./x86/csiplugin ./"
    cp ./x86/csiplugin ./
else 
    echo "cp ./arm/csiplugin ./"
    cp ./arm/csiplugin ./
fi

chmod +x ./csiplugin

cp ./csiplugin /usr/bin/ -rf

cp ./repo/ceph.repo /etc/yum.repos.d/ -rf
