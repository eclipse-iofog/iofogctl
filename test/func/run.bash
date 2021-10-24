#!/usr/bin/env bash

# These functions are designed to be used with the bats `run` command

function curlMsvc(){
    PUBLIC_ENDPOINT="$1"
    curl -s --max-time 120 ${PUBLIC_ENDPOINT}/api/raw
}

function jqMsvcArray(){
    ARR="$1"
    echo "$ARR" | jq '. | length'
}

function findMsvcState(){
    NS="$1"
    MS="$2"
    STATE="$3"
    iofogctl -n $NS get microservices | grep $MS | grep $STATE
}

function runNoExecutors(){
  echo '' > test/conf/nothing.yaml
  iofogctl deploy -f test/conf/nothing.yaml -n "$NS"
}

function runWrongNamespace(){
  echo "---
apiVersion: iofog.org/v3
kind: LocalControlPlane
metadata:
  namespace: wrong
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV
  controller:
    name: func-test" > test/conf/wrong-ns.yaml
  iofogctl deploy -f test/conf/wrong-ns.yaml -n "$NS"
}