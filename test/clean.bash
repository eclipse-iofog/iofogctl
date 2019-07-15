#!/bin/bash

# Export variables
. test/env.sh

echo "Cleaning namespace $NAMESPACE"

iofogctl delete all -n "$NAMESPACE"
kubectl delete all --all -n "$NAMESPACE"
kubectl delete clusterrolebindings kubelet
kubectl delete crd iofogs.k8s.iofog.org
kubectl delete ns "$NAMESPACE"
iofogctl delete namespace "$NAMESPACE"