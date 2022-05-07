#!/bin/bash


HERE=`pwd`

cd redis-6.2.6


make clean

ARCH=`uname -m`
echo `date` $ARCH

malloc="tcmalloc"
echo `date` $malloc

echo `date` make -j"$(nproc)" MALLOC=$malloc redis-server
make -j"$(nproc)" MALLOC=$malloc redis-server

echo `date` make -j"$(nproc)" MALLOC=$malloc redis-sentinel
make -j"$(nproc)" MALLOC=$malloc redis-sentinel

echo `date` make -j"$(nproc)" MALLOC=$malloc redis-cli
make -j"$(nproc)" MALLOC=$malloc redis-cli

echo `date` make -j"$(nproc)" MALLOC=$malloc redis-check-aof
make -j"$(nproc)" MALLOC=$malloc redis-check-aof

echo `date` make -j"$(nproc)" MALLOC=$malloc redis-check-rdb
make -j"$(nproc)" MALLOC=$malloc redis-check-rdb

echo `date` cd $HERE
cd $HERE

echo `date` rm ./redis-server -f
rm ./redis-server -f

echo `date` rm ./redis-sentinel -f
rm ./redis-sentinel -f

echo `date` rm ./redis-cli -f
rm ./redis-cli -f

echo `date` rm ./redis-check-aof -f
rm ./redis-check-aof -f

echo `date` rm ./redis-check-rdb -f
rm ./redis-check-rdb -f

echo `date` cp ./redis-6.2.6/src/redis-server ./ -f
cp ./redis-6.2.6/src/redis-server ./ -f

echo `date` cp ./redis-6.2.6/src/redis-sentinel ./ -f
cp ./redis-6.2.6/src/redis-sentinel ./ -f

echo `date` cp ./redis-6.2.6/src/redis-cli ./ -f
cp ./redis-6.2.6/src/redis-cli ./ -f

echo `date` cp ./redis-6.2.6/src/redis-check-aof ./ -f
cp ./redis-6.2.6/src/redis-check-aof ./ -f

echo `date` cp ./redis-6.2.6/src/redis-check-rdb ./ -f
cp ./redis-6.2.6/src/redis-check-rdb ./ -f

# /usr/bin

echo `date` cp ./redis-server /usr/bin/ -f
cp ./redis-server /usr/bin/ -f

echo `date` cp ./redis-sentinel /usr/bin/ -f
cp ./redis-sentinel /usr/bin/ -f

echo `date` cp ./redis-cli /usr/bin/ -f
cp ./redis-cli /usr/bin/ -f

echo `date` cp ./redis-check-aof /usr/bin/ -f
cp ./redis-check-aof /usr/bin/ -f

echo `date` cp ./redis-check-rdb /usr/bin/ -f
cp ./redis-check-rdb /usr/bin/ -f




