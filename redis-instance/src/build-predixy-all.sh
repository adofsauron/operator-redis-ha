#!/bin/bash

HERE=`pwd`

cd predixy-1.0.5

make clean
make -j"$(nproc)" 

echo `date` cp src/predixy /usr/bin
cp src/predixy /usr/bin

cd $HERE

exit 0