#!/usr/bin/env bash

function kctl(){
  KUBECONFIG="$TEST_KUBE_CONFIG" kubectl $@
}

function print(){
  echo "# $@                                " >&3
}

function log(){
  echo "$@" >> "/tmp/bats_$BATS_TEST_NAME"
}