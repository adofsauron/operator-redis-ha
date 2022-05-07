#!/bin/bash

export ETCDCTL_API=3

echo etcdctl --endpoints="127.0.0.1:2379" --cert=/etc/kubernetes/pki/apiserver-etcd-client.crt --key=/etc/kubernetes/pki/apiserver-etcd-client.key --cacert=/etc/kubernetes/pki/etcd/ca.crt --write-out="table" member list

etcdctl --endpoints="127.0.0.1:2379" --cert=/etc/kubernetes/pki/apiserver-etcd-client.crt --key=/etc/kubernetes/pki/apiserver-etcd-client.key --cacert=/etc/kubernetes/pki/etcd/ca.crt --write-out="table" member list

echo -e "\n"

echo etcdctl --endpoints="127.0.0.1:2379" --cert=/etc/kubernetes/pki/apiserver-etcd-client.crt --key=/etc/kubernetes/pki/apiserver-etcd-client.key --cacert=/etc/kubernetes/pki/etcd/ca.crt  get --prefix / --keys-only=true

etcdctl --endpoints="127.0.0.1:2379" --cert=/etc/kubernetes/pki/apiserver-etcd-client.crt --key=/etc/kubernetes/pki/apiserver-etcd-client.key --cacert=/etc/kubernetes/pki/etcd/ca.crt  get --prefix / --keys-only=true

