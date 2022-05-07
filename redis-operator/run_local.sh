#!/bin/bash

export ETCD_PATH_CERT="/etc/kubernetes/pki/apiserver-etcd-client.crt"
export ETCD_PATH_KEY="/etc/kubernetes/pki/apiserver-etcd-client.key"
export ETCD_PATH_CACERT="/etc/kubernetes/pki/etcd/ca.crt"

./bin/manager

