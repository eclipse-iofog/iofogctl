#!/bin/bash

. test/conf/env.sh

echo "Cleaning namespace $NAMESPACE"

iofogctl delete all -n "$NAMESPACE" -v
kubectl delete controlplanes/iofog -n "$NAMESPACE"
kubectl delete all --all -n "$NAMESPACE"
kubectl delete clusterrolebinding "${NAMESPACE}-iofog-operator"
kubectl delete ns "$NAMESPACE"
iofogctl disconnect -n "$NAMESPACE" -v
iofogctl delete namespace "$NAMESPACE" -v

exit 0
