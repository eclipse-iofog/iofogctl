#!/usr/bin/env bash

function testDeployVolume(){
  SRC="/tmp/iofogctl_tests"
  DST="$VOL_DEST"
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
      SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $SSH_KEY_PATH ${USERS[IDX]}@${HOSTS[IDX]}"
      $SSH_COMMAND -- cat $DST/test$FILE_IDX | grep "test$FILE_IDX"
      $SSH_COMMAND -- cat $DST/testdir/test$FILE_IDX | grep "test$FILE_IDX"
    done
  done
}

function testMountVolume(){
  initAgents
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  SSH_COMMAND="ssh -S -oStrictHostKeyChecking=no -i $SSH_KEY_PATH ${USERS[0]}@${HOSTS[0]}"
  CONTAINER=$($SSH_COMMAND -- sudo docker ps | grep "heart-rate-ui" | awk '{print $1}')
  $SSH_COMMAND -- sudo docker exec $CONTAINER ls $VOL_CONT_DEST
  for FILE_IDX in 1 2 3; do
    $SSH_COMMAND -- sudo docker exec $CONTAINER cat $VOL_CONT_DEST/test$FILE_IDX | grep test$FILE_IDX
    $SSH_COMMAND -- sudo docker exec $CONTAINER cat $VOL_CONT_DEST/testdir/test$FILE_IDX | grep test$FILE_IDX
  done
}