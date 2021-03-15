#!/bin/bash

. test/conf/env.sh

echo "Cleaning namespace $NAMESPACE"

kubectl delete ns "$NAMESPACE"
iofogctl disconnect -n "$NAMESPACE" -v

exit 0
