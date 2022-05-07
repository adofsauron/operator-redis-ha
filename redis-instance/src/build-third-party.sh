#!/bin/bash

HERE=`pwd`

# protobuf

cd third_party
rm protobuf-cpp-3.20.0 -rf
tar -xzvf protobuf-cpp-3.20.0.tar.gz

cd protobuf-cpp-3.20.0/

./configure

make -j"$(nproc)"
make install

cd ..
# rm protobuf-cpp-3.20.0 -rf

# grpc

cd $HERE

cd third_party
rm grpc -rf
tar -xzvf grpc.tar.gz

cd grpc
rm ./build -rf
mkdir -p build
cd build

cmake ..
make -j"$(nproc)"
make install

cd ../..
# rm grpc -rf


# libgo

cd $HERE

cd third_party
rm libgo-3.1-stable -rf
tar -xzvf libgo-3.1-stable.tar.gz

cd libgo-3.1-stable

rm ./build -rf
mkdir -p ./build
cd ./build

cmake -DDISABLE_HOOK=on  ..

make -j"$(nproc)"
make install

cd ../..
rm libgo-3.1-stable -rf

# log4cpp

# cd third_party
# rm log4cpp-2.9.1 -rf
# tar -xzvf log4cpp-2.9.1.tar.gz

# cd log4cpp-2.9.1

# rm ./build -rf
# mkdir -p ./build
# cd ./build

# cmake ..

# make -j"$(nproc)"
# make install

# cd ../..
# rm log4cpp-2.9.1 -rf

# cpprst

cd $HERE

cd third_party
rm cpprestsdk -rf
tar -xzvf cpprestsdk.tar.gz
cd cpprestsdk

rm ./build -rf
mkdir -p ./build
cd ./build

cmake ..

make -j"$(nproc)"
make install

cd ../..
# rm ./cpprestsdk -rf

# etcd-cpp-apiv3

cd $HERE

cd third_party
rm etcd-cpp-apiv3-0.2.3 -rf
tar -xzvf etcd-cpp-apiv3-0.2.3.tar.gz
cd etcd-cpp-apiv3-0.2.3

rm ./build -rf
mkdir -p ./build
cd ./build

cmake ..

make -j"$(nproc)"
make install
cp ./src/libetcd-cpp-api.so  /lib64/ -rf

cd ../..
# rm ./etcd-cpp-apiv3-0.2.3 -rf

# etcd for debug

# cd third_party
# rm etcd-3.6.0-alpha.0 -rf
# tar -xzvf etcd-3.6.0-alpha.0.tar.gz
# cd etcd-3.6.0-alpha.0

# go env -w GOPROXY=https://goproxy.cn,direct
# cd etcdctl
# go mod tidy
# go mod vendor
# go build -o etcdctl ./main.go
# cp ./bin/*  /usr/bin/ -rf

# cd ..

# over 

cd $HERE