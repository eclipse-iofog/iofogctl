#!/usr/bin/env bash

function testDeployVolume(){
  SRC="/tmp/iofogctl_tests"
  DST="/tmp"
  YAML_SRC="$SRC"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    YAML_SRC="C:/tests"
    SRC=$(wslpath $YAML_SRC)
  fi
  initAgents
  echo "---
apiVersion: iofog.org/v1
kind: Volume
spec:
  source: $YAML_SRC
  destination: $DST
  permissions: 666
  agents:
  - $NAME-0
  - $NAME-1" > test/conf/volume.yaml

  [ ! -d $SRC ] && mkdir $SRC
  for IDX in 1 2 3; do
    echo "test$IDX" > "$SRC/test$IDX"
  done
  [ ! -d $SRC/testdir ] && mkdir $SRC/testdir
  for IDX in 1 2 3; do
    echo "test$IDX" > "$SRC/testdir/test$IDX"
  done
  iofogctl -v -n "$NS" deploy -f test/conf/volume.yaml

  # Check files
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    for FILE_IDX in 1 2 3; do
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat $DST/test$FILE_IDX | grep "test$FILE_IDX"
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat $DST/testdir/test$FILE_IDX | grep "test$FILE_IDX"
    done
  done
}