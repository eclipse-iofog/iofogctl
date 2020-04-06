#!/usr/bin/env bash

function testDeployVolume(){
  SRC="$VOL_SRC"
  DST="$VOL_DEST"
  YAML_SRC="$SRC"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    YAML_SRC="$WIN_VOL_SRC"
    SRC=$(wslpath $YAML_SRC)
  fi
  initAgents
  echo "---
apiVersion: iofog.org/v2
kind: Volume
spec:
  name: $VOL_NAME
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

function testGetDescribeVolume(){
  SRC="$VOL_SRC"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SRC="$WIN_VOL_SRC"
  fi

  # Describe
  DESC=$(iofogctl -v -n "$NS" describe volume "$VOL_NAME")
  echo "$DESC"
  [ ! -z "$(echo $DESC | grep $VOL_NAME)" ]
  [ ! -z "$(echo $DESC | grep $NAME-0)" ]
  [ ! -z "$(echo $DESC | grep $NAME-1)" ]
  [ ! -z "$(echo $DESC | grep $SRC)" ]
  [ ! -z "$(echo $DESC | grep $VOL_DEST)" ]
  [ ! -z "$(echo $DESC | grep 666)" ]

  # Get
  GET=$(iofogctl -v -n "$NS" get volumes)
  echo "$GET"
  [ ! -z "$(echo $GET | grep $VOL_NAME)" ]
  [ ! -z "$(echo $GET | grep $NAME-0)" ]
  [ ! -z "$(echo $GET | grep $NAME-1)" ]
  [ ! -z "$(echo $GET | grep $SRC)" ]
  [ ! -z "$(echo $GET | grep $VOL_DEST)" ]
  [ ! -z "$(echo $GET | grep 666)" ]
}

function testDeleteVolume(){
  SRC="$VOL_SRC"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SRC="$WIN_VOL_SRC"
  fi

  iofogctl -v -n "$NS" delete volume "$VOL_NAME"
  GET=$(iofogctl -v -n "$NS" get volumes)
  echo "$GET"
  [ ! -z "$(echo $GET | grep VOLUME)" ]
  [ ! -z "$(echo $GET | grep SOURCE)" ]
  [ -z "$(echo $GET | grep $VOL_NAME)" ]
  [ -z "$(echo $GET | grep $NAME-0)" ]
  [ -z "$(echo $GET | grep $NAME-1)" ]
  [ -z "$(echo $GET | grep $SRC)" ]
  [ -z "$(echo $GET | grep $VOL_DEST)" ]
  [ -z "$(echo $GET | grep 666)" ]

  # Check files
  initAgents
    local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    for FILE_IDX in 1 2 3; do
      SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $SSH_KEY_PATH ${USERS[IDX]}@${HOSTS[IDX]}"
      $SSH_COMMAND -- [ -z "$(ls $VOL_DEST | xargs echo)" ]
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

function testDefaultProxyConfig(){
  ACTUAL_IP="$1"
  ITER=0
  IP=""
  while [ $ITER -lt 10 ] && [ "$IP" != "$ACTUAL_IP" ]; do
    sleep 6
    IP=$(iofogctl -n "$NS" -v describe microservice "$MSVC2_NAME" | grep publicLink | sed 's/.*http:\/\///g' | sed 's/:.*//g')
    ITER=$((ITER+1))
  done
  echo "Found IP: $IP"
  echo "Wanted IP: $ACTUAL_IP"
  [ "$ACTUAL_IP" == "$IP" ]
}
