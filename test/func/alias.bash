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

function startTest(){
  local FILE="/tmp/bats.test"
  if [ -f $FILE ]; then
    skip
  fi
  echo 'start' > $FILE
}

function stopTest(){
  local FILE="/tmp/bats.test"
  if [ -f $FILE ]; then
    rm $FILE
  fi
}