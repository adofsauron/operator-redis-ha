#!/bin/bash

kubectl apply -f config/samples/apps_v1alpha1_operatorredisha.yaml



# kubectl taint node node-201 node.kubernetes.io/disk-pressure-

# kubectl  describe pod operatorredisha-sample-0

# kubectl  describe node node-201  | grep Tain

