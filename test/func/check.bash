#!/usr/bin/env bash

function checkControllerK8s {
  # TODO: Replace this one controller pod name is returned
  OLD_NAME="$NAME"
  NAME="$1"
  checkController
  NAME="$OLD_NAME"
}

function checkControllerNegativeK8s {
  # TODO: Replace this one controller pod name is returned
  OLD_NAME="$NAME"
  NAME="$1"
  checkControllerNegative
  NAME="$OLD_NAME"
}

function checkController() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe controller "$NAME" | grep "name: $NAME") ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe controlplane | grep "ame: $NAME") ]]
}

function checkControllerNegative() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkMicroservice() {
  NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe microservice "$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  MSVC_CONFIG=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $4}')
  checkMsvcConfig "${MSVC_CONFIG}" "\"test_mode\":true"
  checkMsvcConfig "${MSVC_CONFIG}" "\"data_label\":\"Anonymous_Person_2\""
  [[ "memoryLimit: 8192" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep memoryLimit | awk '{$1=$1};1' ) ]]
  # Check route
  [[ "$MSVC1_NAME, $MSVC2_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk -F '\t' '{print $6}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "5005:443")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "/tmp/microservice:/tmp")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS_CHECK" describe microservice "$MICROSERVICE_NAME" -o "test/conf/msvc_output.yaml"
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "name: $MICROSERVICE_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5005") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- hostDestination: /tmp/microservice") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "containerDestination: /tmp") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:test") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:test-arm") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"42\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "data_label: Anonymous_Person_2") ]]
}

function checkUpdatedMicroservice() {
  NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe microservice "$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  MSVC_CONFIG=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $4}')
  checkMsvcConfig "${MSVC_CONFIG}" "\"test_mode\":true"
  checkMsvcConfig "${MSVC_CONFIG}" "\"data_label\":\"Anonymous_Person_3\""
  checkMsvcConfig "${MSVC_CONFIG}" "\"test_data\":{\"key\":42}"
  [[ "memoryLimit: 5555" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep memoryLimit | awk '{$1=$1};1' ) ]]
  [[ "diskDirectory: /tmp/iofog-agent/" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep diskDirectory | awk '{$1=$1};1') ]]
  # Check route
  [[ "$MSVC1_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk -F '\t' '{print $6}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "5443:443, 5080:80")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "/tmp/updatedmicroservice:/tmp")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS_CHECK" describe microservice "$MICROSERVICE_NAME" -o "test/conf/msvc_output.yaml"
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "name: $MICROSERVICE_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC1_NAME") ]]
  [[ -z $(cat test/conf/msvc_output.yaml | grep "\- $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 443") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "external: 5080") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- internal: 80") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- hostDestination: /tmp/updatedmicroservice") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "containerDestination: /tmp") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:test") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:test-arm") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"75\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "\- key: TEST_2") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "value: \"42\"") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "test_data:") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "key: 42") ]]
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "data_label: Anonymous_Person_3") ]]
}

function checkMicroserviceNegative() {
  NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
}

# Takes the config as $1 and the expected key:value as $2
function checkMsvcConfig() {
  [[ ! -z $(echo $1 | grep $2) ]]
}

function checkApplication() {
  NS_CHECK=${1:-$NS}
  iofogctl -v -n "$NS_CHECK" get applications
  [[ "$APPLICATION_NAME" == $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  MSVCS=$(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" )
  echo "$MSVCS" | grep "$MSVC1_NAME"
  echo "$MSVCS" | grep "$MSVC2_NAME"
  MSVCS=$(iofogctl -v -n "$NS_CHECK" get microservices)
  echo "$MSVCS" | grep "$MSVC1_NAME"
  echo "$MSVCS" | grep "$MSVC2_NAME"

  # Check config
  MSVC_CONFIG=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MSVC1_NAME" | awk '{print $4}')
  checkMsvcConfig "${MSVC_CONFIG}" "\"test_mode\":true"
  checkMsvcConfig "${MSVC_CONFIG}" "\"data_label\":\"Anonymous_Person\""
  [[ "bluetoothEnabled: true" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep bluetooth | awk '{$1=$1};1' ) ]]
  # Check route
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MSVC1_NAME" | awk '{print $5}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "5000:80")
  [[ "$MSVC2_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "/tmp/msvc:/tmp")
  [[ "$MSVC1_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS_CHECK" describe application "$APPLICATION_NAME" -o "test/conf/app_output.yaml"
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $APPLICATION_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- from: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "to: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "external: 5000") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- internal: 80") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- hostDestination: /tmp/msvc") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "containerDestination: /tmp") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:x86-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:arm-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "\- key: BASE_URL") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "value: http://localhost:8080/data") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "data_label: Anonymous_Person") ]]
}

function checkApplicationNegative() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ "$MSVC1_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  [[ "$MSVC2_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MSVC2_NAME" | awk '{print $1}') ]]
}

function checkAgent() {
  NS_CHECK=${2:-$NS}
  OPTIONS=$3
  AGENT_NAME=$1
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS_CHECK" get agents $OPTIONS | grep "$AGENT_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe agent "$AGENT_NAME" $OPTIONS | grep "name: $AGENT_NAME") ]]
}

function checkDetachedAgent() {
  AGENT_NAME=$1
  NS_CHECK=${2:-$NS}
  # Check agent is accessible using ssh, and is not provisioned
  [[ "not" == $(iofogctl -v legacy agent $AGENT_NAME status --detached | grep 'Connection to Controller' | awk '{print $5}') ]]
  # Check agent is listed in detached resources
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS_CHECK" get agents --detached | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkDetachedAgentNegative() {
  AGENT_NAME=$1
  # Check agent is not listed in detached resources
  [[ "$AGENT_NAME" != $(iofogctl -v get agents --detached | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgentNegative() {
  NS_CHECK=${2:-$NS}
  AGENT_NAME=$1
  [[ "$AGENT_NAME" != $(iofogctl -v -n "$NS_CHECK" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgents() {
  NS_CHECK=${1:-$NS}
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    checkAgent "$AGENT_NAME" "$NS_CHECK"
  done
}

function checkAgentsNegative() {
  NS_CHECK=${1:-$NS}
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    checkAgentNegative "$AGENT_NAME" "$NS_CHECK"
  done
}

function checkAgentListFromController() {
  local API_ENDPOINT=$(cat /tmp/api_endpoint.txt)
  local ACCESS_TOKEN=$(cat /tmp/access_token.txt)
  local LIST=$(curl --request GET \
--url $API_ENDPOINT/api/v3/iofog-list \
--header "Authorization: $ACCESS_TOKEN" \
--header 'Content-Type: application/json')
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    local UUID=$(echo $LIST | jq -r '.fogs[] | select(.name == "'"$AGENT_NAME"'") | .uuid')
    [[ ! -z "$UUID" ]]
  done
}

function checkAgentPruneController(){
  local API_ENDPOINT="$1"
  local KEY_FILE="$2"
  local AGENT_TOKEN=$(ssh -oStrictHostKeyChecking=no -i $KEY_FILE ${USERS[0]}@${HOSTS[0]} -- cat /etc/iofog-agent/config.xml  | grep 'access_token' | tr -d '<' | tr -d '/' | tr -d '>' | awk -F 'access_token' '{print $2}')
  local CHANGES=$(curl --request GET \
--url $API_ENDPOINT/api/v3/agent/config/changes \
--header "Authorization: $AGENT_TOKEN" \
--header 'Content-Type: application/json')
  local PRUNE=$(echo $CHANGES | jq -r .prune)
  [[ "true" == "$PRUNE" ]]
}

function checkLegacyController() {
  NS_CHECK=${1:-$NS}
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" legacy controller $NAME status | grep 'ioFogController') ]]
}

function checkLegacyAgent() {
  NS_CHECK=${2:-$NS}
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" legacy agent $1 status | grep 'RUNNING') ]]
  [[ "ok" == $(iofogctl -v -n "$NS_CHECK" legacy agent $1 status | grep 'Connection to Controller' | awk '{print $5}') ]]
}

function checkMovedMicroservice() {
  MSVC=$1
  NEW_AGENT=$2
  [[ ! -z $(iofogctl -v get microservices | grep $MSVC | grep $NEW_AGENT) ]]
}

function checkRenamedResource() {
  RSRC=$1
  OLDNAME=$2
  NEWNAME=$3
  NAMESPACE=$4
  [[ -z $(iofogctl -n ${NAMESPACE} -v get ${RSRC} | grep -w ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -n ${NAMESPACE} -v get ${RSRC} | grep -w ${NEWNAME}) ]]
}

function checkRenamedApplication() {
  OLDNAME=$1
  NEWNAME=$2
  NAMESPACE=$3

  [[ -z $(iofogctl -n ${NAMESPACE} -v get applications | awk '{print $1}' | grep ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -n ${NAMESPACE} -v get applications |  awk '{print $1}' | grep ${NEWNAME}) ]]
}

function checkNamespaceExistsNegative() {
  CHECK_NS="$1"
  [ -z "$(iofogctl get namespaces | grep $CHECK_NS)" ]
}

function checkRenamedNamespace() {
  OLDNAME=$1
  NEWNAME=$2
  [[ -z $(iofogctl -v get namespaces | grep -w ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -v get namespaces | grep -w ${NEWNAME}) ]]
}

function hitMsvcEndpoint() {
  IP="$1"
  ITER=0
  COUNT=0
  while [ $COUNT -eq 0 ] && [ $ITER -lt 24 ]; do
    sleep 10
    run curlMsvc "$IP"
    if [ $status -eq 0 ]; then
      RET="$output"
      echo "$RET"
  
      run jqMsvcArray "$RET"
      if [ $status -eq 0 ]; then
        COUNT=$output
      fi
    fi
    ITER=$((ITER+1))
  done
  [ $COUNT -gt 0 ]
}

function checkVanillaResourceDeleted() {
  USER=$1
  HOST=$2
  PORT=$3
  KEY_FILE=$4
  RESOURCE=$5

  [[ -z $(ssh -oStrictHostKeyChecking=no $USER@$HOST:$PORT -i $KEY_FILE sudo which ${RESOURCE}) ]]
}

function checkLocalResourcesDeleted() {
  docker ps -a
  [[ -z $(docker ps -aq) ]]
}

function checkGCRRegistry() {
  iofogctl get -n "$NS" registries
  iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $1}'
  iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $2}'
  [[ "3" == $(iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $1}') ]]
  [[ "gcr.io" == $(iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $2}') ]]
}

function checkUpdatedGCRRegistry() {
  iofogctl get -n "$NS" registries
  iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $1}'
  iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $2}'
  [[ "3" == $(iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $1}') ]]
  [[ "https://gcr.io" == $(iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $2}') ]]
}

function checkGCRRegistryNegative() {
  [[ -z $(iofogctl get -n "$NS" registries | grep gcr.io | awk '{print $1}') ]]
}
