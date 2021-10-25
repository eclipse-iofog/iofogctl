#!/usr/bin/env bash

function testDeployLocalVolume(){
  SRC="$VOL_DEST"
  DST="$VOL_CONT_DEST"
  YAML_SRC="$SRC"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    YAML_SRC="$WIN_VOL_SRC"
    SRC=$(wslpath $YAML_SRC)
  fi
  echo "---
apiVersion: iofog.org/v3
kind: Volume
spec:
  name: $VOL_NAME
  source: $YAML_SRC
  destination: $DST
  permissions: 666
  agents:
  - $NAME-0" > test/conf/volume.yaml

  [ ! -d $SRC ] && mkdir $SRC
  for IDX in 1 2 3; do
    echo "test$IDX" > "$SRC/test$IDX"
  done
  for DIR_IDX in 1 2 3; do
    [ ! -d $SRC/testdir$DIR_IDX ] && mkdir $SRC/testdir$DIR_IDX
    for FILE_IDX in 1 2 3; do
      echo "test$FILE_IDX" > "$SRC/testdir$DIR_IDX/test$FILE_IDX"
    done
  done
  iofogctl -v -n "$NS" deploy -f test/conf/volume.yaml
}

function testGetDescribeLocalVolume(){
  SRC="$VOL_DEST"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SRC=wsl "$SRC"
  fi

  # Describe
  DESC=$(iofogctl -v -n "$NS" describe volume "$VOL_NAME")
  echo "$DESC"
  [ ! -z "$(echo $DESC | grep $VOL_NAME)" ]
  [ ! -z "$(echo $DESC | grep $NAME-0)" ]
  [ ! -z "$(echo $DESC | grep $SRC)" ]
  [ ! -z "$(echo $DESC | grep $VOL_DEST)" ]
  [ ! -z "$(echo $DESC | grep 666)" ]

  # Get
  GET=$(iofogctl -v -n "$NS" get volumes)
  echo "$GET"
  [ ! -z "$(echo $GET | grep $VOL_NAME)" ]
  [ ! -z "$(echo $GET | grep $NAME-0)" ]
  [ ! -z "$(echo $GET | grep $SRC)" ]
  [ ! -z "$(echo $GET | grep $VOL_DEST)" ]
  [ ! -z "$(echo $GET | grep 666)" ]
}

function testDeleteLocalVolume(){
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
  [ -z "$(echo $GET | grep $SRC)" ]
  [ -z "$(echo $GET | grep $VOL_DEST)" ]
  [ -z "$(echo $GET | grep 666)" ]
}

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
apiVersion: iofog.org/v3
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
  for DIR_IDX in 1 2 3; do
    [ ! -d $SRC/testdir$DIR_IDX ] && mkdir $SRC/testdir$DIR_IDX
    for FILE_IDX in 1 2 3; do
      echo "test$FILE_IDX" > "$SRC/testdir$DIR_IDX/test$FILE_IDX"
    done
  done
  iofogctl -v -n "$NS" deploy -f test/conf/volume.yaml

  # Check files
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    for DIR_IDX in 1 2 3; do
      for FILE_IDX in 1 2 3; do
        SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $SSH_KEY_PATH ${USERS[IDX]}@${HOSTS[IDX]}"
        $SSH_COMMAND -- cat $DST/test$FILE_IDX
        $SSH_COMMAND -- cat $DST/test$FILE_IDX | grep "test$FILE_IDX"
        $SSH_COMMAND -- cat $DST/testdir$DIR_IDX/test$FILE_IDX
        $SSH_COMMAND -- cat $DST/testdir$DIR_IDX/test$FILE_IDX | grep "test$FILE_IDX"
      done
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
      RESULT=$($SSH_COMMAND -- ls $VOL_DEST)
      [ -z "$RESULT" ]
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
  for DIR_IDX in 1 2 3; do
    for FILE_IDX in 1 2 3; do
      $SSH_COMMAND -- sudo docker exec $CONTAINER cat $VOL_CONT_DEST/test$FILE_IDX | grep test$FILE_IDX
      $SSH_COMMAND -- sudo docker exec $CONTAINER cat $VOL_CONT_DEST/testdir$DIR_IDX/test$FILE_IDX | grep test$FILE_IDX
    done
  done
}

function testDefaultProxyConfig(){
  ACTUAL_IP="$1"
  ITER=0
  IP=""
  while [ $ITER -lt 10 ] && [ "$IP" != "$ACTUAL_IP" ]; do
    sleep 6
    IP=$(iofogctl -n "$NS" -v describe microservice $APPLICATION_NAME/"$MSVC2_NAME" | grep "\- http://" | sed 's/.*http:\/\///g' | sed 's/:.*//g')
    ITER=$((ITER+1))
  done
  echo "Found IP: $IP"
  echo "Wanted IP: $ACTUAL_IP"
  [ "$ACTUAL_IP" == "$IP" ]
}

function testNoExecutors(){
  run runNoExecutors
  [ $status -ne 0 ]
  echo "$output"
  [[ "$output" == *"not decode any valid resources"* ]]
}

function testWrongNamespace(){
  run runWrongNamespace
  [ $status -ne 0 ]
  echo "$output"
  [[ "$output" == *"does not match the Namespace"* ]]
}

function testDefaultNamespace(){
  local SET_NS="$1"
  iofogctl configure current-namespace "$NS"
  iofogctl get all | grep "$NS"
}

function testGenerateConnectionString(){
  local ADDR="$1"
  local CNCT=$(iofogctl -n "$NS" connect --generate)
  echo "Wanted: $CNCT"
  echo "Got: iofogctl connect --ecn-addr $ADDR --name remote --email $USER_EMAIL --pass $USER_PW_B64 --b64"
  [ "$CNCT" == "iofogctl connect --ecn-addr $ADDR --name remote --email $USER_EMAIL --pass $USER_PW_B64 --b64" ]
}

function testEdgeResources(){
  initEdgeResourceFile
  initAgents

  # Create first version
  iofogctl -n "$NS" deploy -f test/conf/edge-resource.yaml
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_NAME)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_PROTOCOL)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep "$EDGE_RESOURCE_DESC")" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep $EDGE_RESOURCE_NAME)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep $EDGE_RESOURCE_PROTOCOL)" ]
  # Test idempotence
  iofogctl -n "$NS" deploy -f test/conf/edge-resource.yaml

  # Attach first version
  local AGENT="${NAME}-0"
  iofogctl -n "$NS" attach edge-resource "$EDGE_RESOURCE_NAME" "$EDGE_RESOURCE_VERSION" "$AGENT"
  [ ! -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- smart")" ]
  [ ! -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- door")" ]

  # Detach first version
  iofogctl -n "$NS" detach edge-resource "$EDGE_RESOURCE_NAME" "$EDGE_RESOURCE_VERSION" "$AGENT"
  [ -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- smart")" ]
  [ -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- door")" ]

  # Deploy new version
  local ER_VERS='v1.0.1'
  initEdgeResourceFile "$ER_VERS"
  iofogctl -n "$NS" deploy -f test/conf/edge-resource.yaml
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $ER_VERS)" ]

  # Attach new version
  iofogctl -n "$NS" attach edge-resource "$EDGE_RESOURCE_NAME" "$EDGE_RESOURCE_VERSION" "$AGENT"
  [ ! -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- smart")" ]
  [ ! -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- door")" ]

  # Rename
  local NEW_NAME="smart-car"
  iofogctl -n "$NS" rename edge-resource "$EDGE_RESOURCE_NAME" "$NEW_NAME"
  [ -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_NAME)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $NEW_NAME)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_PROTOCOL)" ]
  [ -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep $EDGE_RESOURCE_NAME)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource "$NEW_NAME" "$EDGE_RESOURCE_VERSION" | grep "$EDGE_RESOURCE_DESC")" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource "$NEW_NAME" "$EDGE_RESOURCE_VERSION" | grep $NEW_NAME)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource "$NEW_NAME" "$EDGE_RESOURCE_VERSION" | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource "$NEW_NAME" "$EDGE_RESOURCE_VERSION" | grep $EDGE_RESOURCE_PROTOCOL)" ]
  iofogctl -n "$NS" rename edge-resource "$NEW_NAME" "$EDGE_RESOURCE_NAME"
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_NAME)" ]
  [ ! -z "$(iofogctl -n $NS describe edge-resource $EDGE_RESOURCE_NAME $EDGE_RESOURCE_VERSION | grep $EDGE_RESOURCE_NAME)" ]

  # Delete both versions
  iofogctl -n "$NS" delete edge-resource "$EDGE_RESOURCE_NAME" "$EDGE_RESOURCE_VERSION"
  [ -z "$(iofogctl -n $NS get edge-resources | grep $EDGE_RESOURCE_VERSION)" ]
  [ ! -z "$(iofogctl -n $NS get edge-resources | grep $ER_VERS)" ]
  iofogctl -n "$NS" delete edge-resource "$EDGE_RESOURCE_NAME" "$ER_VERS"
  [ -z "$(iofogctl -n $NS get edge-resources | grep $ER_VERS)" ]
  [ -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- smart")" ]
  [ -z "$(iofogctl -n $NS describe agent $AGENT | grep "\- door")" ]
}

function testApplicationTemplates(){
  initApplicationTemplateFile
  initAgents

  # Deploy and verify
  iofogctl -v -n "$NS" deploy -f test/conf/app-template.yaml
  for CHECK in "$APP_TEMPLATE_NAME" "$APP_TEMPLATE_DESC" "2" "1"; do
    iofogctl -v -n "$NS" get application-templates | grep "$CHECK"
  done
  for CHECK in "123" "bobbing" "pineapple" "6666" "{{public-port}}"; do
    echo "Wanted $CHECK"
    iofogctl -v -n "$NS" describe application-template "$APP_TEMPLATE_NAME" | grep "$CHECK"
  done

  # Delete and verify
  iofogctl -v -n "$NS" delete application-template "$APP_TEMPLATE_NAME"
  [ -z "$(iofogctl -v -n "$NS" get application-templates | grep "$APP_TEMPLATE_NAME")" ]

  # Deploy again and deploy application
  iofogctl --debug -n "$NS" deploy -f test/conf/app-template.yaml
  iofogctl --debug -n "$NS" deploy -f test/conf/templated-app.yaml
  checkApplication "$NS" 80 7777 6666

  # Look for templated variables
  for CHECK in 12345 7777 80 func-test-0 "rootHostAccess: false"; do
    echo "Wanted $CHECK"
    iofogctl -v -n "$NS" describe application "$APPLICATION_NAME" | grep "$CHECK"
  done

  # Delete templated app and template
  iofogctl -v -n "$NS" delete -f test/conf/templated-app.yaml
  iofogctl -v -n "$NS" delete application-template "$APP_TEMPLATE_NAME"
}

# testAgentCount enforces a minimum of 2 Agents to allow for regular and custom installs to be tested simultaneously
function testAgentCount(){
  initAgents
  local AGENT_COUNT=${#AGENTS[@]}
  [ "$AGENT_COUNT" -gt 1 ]
}