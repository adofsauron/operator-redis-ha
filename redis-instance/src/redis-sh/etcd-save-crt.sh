#!/bin/bash

BASE64_ETCD_CERT=$1
BASE64_ETCD_KEY=$2
BASE64_ETCD_CACER=$3

echo `date` BASE64_ETCD_CERT=$BASE64_ETCD_CERT
echo `date` BASE64_ETCD_KEY=$BASE64_ETCD_KEY
echo `date` BASE64_ETCD_CACER=$BASE64_ETCD_CACER

ETCD_CERT=`echo $BASE64_ETCD_CERT | base64 -d`
ETCD_KEY=`echo $BASE64_ETCD_KEY | base64 -d`
ETCD_CACER=`echo $BASE64_ETCD_CACER | base64 -d`

mkdir -p /etc/etcd
echo -e "$ETCD_CERT" > /etc/etcd/etcd.cert
echo -e "$ETCD_KEY" > /etc/etcd/etcd.key
echo -e "$ETCD_CACER" > /etc/etcd/etcd.cacer

exit 0
