#!/usr/bin/env bash

function waitForK8sController {
  local NS_CHECK=${1:-$NS}
  # Pod will restart once, so wait for either of them to be ready
  kubectl -n "$NS_CHECK" wait --for=condition=Ready pods -l name=controller --timeout 1m || kubectl -n "$NS_CHECK" wait --for=condition=Ready pods -l name=controller --timeout 1m
  kubectl -n "$NS_CHECK" wait --for=condition=Ready pods -l name=port-manager --timeout 1m
  kubectl -n "$NS_CHECK" wait --for=condition=Ready pods -l name=router --timeout 1m
}

function checkControllerK8s {
  local NS_CHECK=${1:-$NS}
  waitForK8sController "$NS_CHECK"
  for NAME in $(kubectl get pods -n "$NS_CHECK" | grep controller | awk '{print $1}'); do
    [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
    iofogctl -v -n "$NS_CHECK" describe controller "$NAME" | grep "name: $NAME"

    local DESC=$(iofogctl -v -n "$NS_CHECK" describe controlplane)
    echo "$DESC" | grep "podName: $NAME"
    echo "$DESC" | grep "config:.*$(echo $KUBE_CONFIG | tr -d '~')"
    echo "$DESC" | grep "kind: KubernetesControlPlane"

    DESC=$(iofogctl -v -n "$NS_CHECK" describe controller "$NAME")
    echo "$DESC" | grep "podName: $NAME"
    echo "$DESC" | grep "kind: KubernetesController"
  done
}

function checkControllerNegativeK8s {
  local NAME="$1"
  local NS_CHECK=${2:-$NS}
  for NAME in $(kubectl get pods -n "$NS_CHECK" | grep controller | awk '{print $1}'); do
    [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
  done
}

function checkControllerAfterConfigure() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]

  local DESC=$(iofogctl -v -n "$NS_CHECK" describe controller "$NAME")
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "user: $VANILLA_USER"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "port: $VANILLA_PORT"
  echo "$DESC" | grep "keyFile:.*$(echo $KEY_FILE | tr -d '~')"
  echo "$DESC" | grep "kind: Controller"

  DESC=$(iofogctl -v -n "$NS_CHECK" describe controlplane)
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "user: $VANILLA_USER"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "port: $VANILLA_PORT"
  echo "$DESC" | grep "keyFile:.*$(echo $KEY_FILE | tr -d '~')"
  echo "$DESC" | grep "kind: ControlPlane"
}

function checkControllerAfterConnect() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]

  local DESC=$(iofogctl -v -n "$NS_CHECK" describe controller "$NAME")
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "kind: Controller"

  DESC=$(iofogctl -v -n "$NS_CHECK" describe controlplane)
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "kind: ControlPlane"
}

function checkController() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]

  local DESC=$(iofogctl -v -n "$NS_CHECK" describe controller "$NAME")
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "user: $VANILLA_USER"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "port: $VANILLA_PORT"
  echo "$DESC" | grep "keyFile:.*$(echo $KEY_FILE | tr -d '~')"
  echo "$DESC" | grep "kind: Controller"

  DESC=$(iofogctl -v -n "$NS_CHECK" describe controlplane)
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "user: $VANILLA_USER"
  echo "$DESC" | grep "host: $VANILLA_HOST"
  echo "$DESC" | grep "port: $VANILLA_PORT"
  echo "$DESC" | grep "keyFile:.*$(echo $KEY_FILE | tr -d '~')"
  [ ! -z "$CONTROLLER_REPO" ] && echo "$DESC" | grep "repo: $CONTROLLER_REPO"
  [ ! -z "$CONTROLLER_VANILLA_VERSION" ] && echo "$DESC" | grep "version: $CONTROLLER_VANILLA_VERSION"
  [ ! -z "$CONTROLLER_PACKAGE_CLOUD_TOKEN" ] && echo "$DESC" | grep "token: $CONTROLLER_PACKAGE_CLOUD_TOKEN"
  [ ! -z "$AGENT_REPO" ] && echo "$DESC" | grep "repo: $AGENT_REPO"
  [ ! -z "$AGENT_VANILLA_VERSION" ] && echo "$DESC" | grep "version: $AGENT_VANILLA_VERSION"
  [ ! -z "$AGENT_PACKAGE_CLOUD_TOKEN" ] && echo "$DESC" | grep "token: $AGENT_PACKAGE_CLOUD_TOKEN"
  echo "$DESC" | grep "email: $USER_EMAIL"
  echo "$DESC" | grep "password: $USER_PW_B64"
  echo "$DESC" | grep "kind: ControlPlane"
}

function checkControllerLocal() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]

  local DESC=$(iofogctl -v -n "$NS_CHECK" describe controller "$NAME")
  echo "$DESC"
  echo "$DESC" | grep "name: $NAME"
  [ ! -z "$CONTROLLER_IMAGE" ] && echo "$DESC" | grep "image: $CONTROLLER_IMAGE"

  DESC=$(iofogctl -v -n "$NS_CHECK" describe controlplane)
  echo "$DESC" | grep "name: $NAME"
  [ ! -z "$CONTROLLER_IMAGE" ] && echo "$DESC" | grep "image: $CONTROLLER_IMAGE"
  echo "$DESC" | grep "email: $USER_EMAIL"
  echo "$DESC" | grep "password: $USER_PW_B64"
  echo "$DESC" | grep "kind: LocalControlPlane"
}

function checkControllerNegative() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkMicroservice() {
  local NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  DESC_MSVC=$(iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"$MICROSERVICE_NAME")
  echo "${DESC_MSVC}" | grep "test_mode: true"
  echo "${DESC_MSVC}" | grep "data_label: Anonymous_Person_2"
  [[ "memoryLimit: 8192" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep memoryLimit | awk '{$1=$1};1' ) ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "5005:443")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "/tmp/microservice:/tmp")
  [[ "$MICROSERVICE_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"$MICROSERVICE_NAME" -o "test/conf/msvc_output.yaml"
  [[ ! -z $(cat test/conf/msvc_output.yaml | grep "name: $APPLICATION_NAME/$MICROSERVICE_NAME") ]]
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
  local NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"$MICROSERVICE_NAME" | grep "name: $MICROSERVICE_NAME") ]]
  # Check config
  DESC_MSVC=$(iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"$MICROSERVICE_NAME")
  echo "${DESC_MSVC}" | grep "test_mode: true"
  echo "${DESC_MSVC}" | grep "data_label: Anonymous_Person_3"
  echo "${DESC_MSVC}" | grep "test_data:"
  echo "${DESC_MSVC}" | grep "key: 42"
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
  local NS_CHECK=${1:-$NS}
  [[ "$MICROSERVICE_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MICROSERVICE_NAME" | awk '{print $1}') ]]
}

function checkApplication() {
  local NS_CHECK=${1:-$NS}
  local PORT_INT=${2:-80}
  local PORT_EXT=${3:-5000}
  iofogctl -v -n "$NS_CHECK" get applications
  [[ "$APPLICATION_NAME" == $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  MSVCS=$(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" )
  echo "$MSVCS" | grep "${MSVC1_NAME}"
  echo "$MSVCS" | grep "${MSVC2_NAME}"
  MSVCS=$(iofogctl -v -n "$NS_CHECK" get microservices)
  echo "$MSVCS" | grep "${MSVC1_NAME}"
  echo "$MSVCS" | grep "${MSVC2_NAME}"

  # Check config
  DESC_MSVC=$(iofogctl -v -n "$NS_CHECK" describe microservice $APPLICATION_NAME/"${MSVC1_NAME}")
  echo "${DESC_MSVC}" | grep "test_mode: true"
  echo "${DESC_MSVC}" | grep "data_label: Anonymous_Person"
  # Deploying an application should not update Agent config. This is a legacy overload of the functionality.
  # Use a separate AgentConfig Kind to update the required config.
  # [[ "bluetoothEnabled: true" == $(iofogctl -v -n "$NS_CHECK" describe agent-config "${NAME}-0" | grep bluetooth | awk '{$1=$1};1' ) ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "$PORT_EXT:$PORT_INT")
  [[ "${MSVC2_NAME}" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS_CHECK" get microservices | grep "$VOL_DEST:$VOL_CONT_DEST")
  [[ "${MSVC2_NAME}" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS_CHECK" describe application "$APPLICATION_NAME" -o "test/conf/app_output.yaml"
  cat test/conf/app_output.yaml | grep "name: $APPLICATION_NAME"
  cat test/conf/app_output.yaml | grep "name: ${MSVC1_NAME}"
  cat test/conf/app_output.yaml | grep "name: ${MSVC2_NAME}"
  cat test/conf/app_output.yaml | grep "ports:"
  cat test/conf/app_output.yaml | grep "external: $PORT_EXT"
  cat test/conf/app_output.yaml | grep "internal: $PORT_INT"
  cat test/conf/app_output.yaml | grep "links:"
  cat test/conf/app_output.yaml | grep "volumes: \[\]"
  cat test/conf/app_output.yaml | grep "images:"
  cat test/conf/app_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:x86-v1"
  cat test/conf/app_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:arm-v1"
  cat test/conf/app_output.yaml | grep "env:"
  cat test/conf/app_output.yaml | grep "\- key: BASE_URL"
  cat test/conf/app_output.yaml | grep "value: http://localhost:8080/data"
  cat test/conf/app_output.yaml | grep "config:"
  cat test/conf/app_output.yaml | grep "test_mode: true"
  cat test/conf/app_output.yaml | grep "data_label: Anonymous_Person"
}

function checkApplicationNegative() {
  local NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ "$MSVC1_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "${MSVC1_NAME}" | awk '{print $1}') ]]
  [[ "$MSVC2_NAME" != $(iofogctl -v -n "$NS_CHECK" get microservices | grep "${MSVC2_NAME}" | awk '{print $1}') ]]
}

function checkPullPercentageOfMicroservice() {
  local MS=$1
  local NS=$2
  [[ ! -z $(iofogctl get microservices -n "$NS"  | grep "%") ]]
  # iofogctl -n "$NS" get microservices | grep "$MS" | awk '{print $3}' |  sed -E "s/.*\((.*)\%\).*/\1/g"
  [[ ! -z $(iofogctl -n "$NS" get microservices | grep "$MS" | awk '{print $3}' |  sed -E "s/.*\((.*)\%\).*/\1/g") ]]
  sleep 2
  PERCENTAGE1=$(iofogctl -n "$NS" get microservices | grep "$MS" | awk '{print $3}' |  sed -E "s/.*\((.*)\%\).*/\1/g")
  sleep 2
  PERCENTAGE2=$(iofogctl -n "$NS" get microservices | grep "$MS" |  awk '{print $3}' |  sed -E "s/.*\((.*)\%\).*/\1/g")
  echo "$PERCENTAGE1"
  echo "$PERCENTAGE2"
  PERCENTAGE1=$((PERCENTAGE1+0))
  PERCENTAGE2=$((PERCENTAGE2+0))
  echo "$PERCENTAGE1"
  echo "$PERCENTAGE2"
  # [[ ! -z $(iofogctl -n "$NS" get microservices | grep "$MS" | awk '{print $3}' |  sed -E "s/.*\((.*)\%\).*/\1/g") ]]
  [ $PERCENTAGE2 -eq $PERCENTAGE1 ] || [ $PERCENTAGE2 -gt $PERCENTAGE1 ]
}


function checkAgent() {
  local NS_CHECK=${2:-$NS}
  local OPTIONS=$3
  local AGENT_NAME=$1
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS_CHECK" get agents $OPTIONS | grep "$AGENT_NAME" | awk '{print $1}') ]]
  # Checks list taken from init.bash
  CHECKS=("name:.$AGENT_NAME"
"description:.special.test.agent"
"bluetoothEnabled:.true"
"abstractedHardwareEnabled:.false"
"latitude:.46\.464646"
"longitude:.64\.646464"
"memoryLimit:.8192"
"diskDirectory:./tmp/iofog-agent/"
"diskLimit:.77"
"cpuLimit:.89"
"logLimit:.12"
"logFileCount:.11"
"statusFrequency:.9"
"changeFrequency:.8"
"deviceScanFrequency:.61")
  for CHECK in ${CHECKS[@]}; do
    echo $CHECK
    [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe agent "$AGENT_NAME" $OPTIONS | grep "$CHECK") ]]
  done
}

function checkDetachedAgent() {
  local AGENT_NAME=$1
  local NS_CHECK=${2:-$NS}
  # Check agent is accessible using ssh, and is not provisioned
  [[ "not" == $(iofogctl -v legacy agent $AGENT_NAME status --detached | grep 'Connection to Controller' | awk '{print $5}') ]]
  # Check agent is listed in detached resources
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS_CHECK" get agents --detached | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkDetachedAgentNegative() {
  local AGENT_NAME=$1
  # Check agent is not listed in detached resources
  [[ "$AGENT_NAME" != $(iofogctl -v get agents --detached | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgentNegative() {
  local NS_CHECK=${2:-$NS}
  local AGENT_NAME=$1
  [[ "$AGENT_NAME" != $(iofogctl -v -n "$NS_CHECK" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgents() {
  local NS_CHECK=${1:-$NS}
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-$(((IDX++)))"
    checkAgent "$AGENT_NAME" "$NS_CHECK"
  done
}

function checkAgentsNegative() {
  local NS_CHECK=${1:-$NS}
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
  local NS_CHECK=${1:-$NS}
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" legacy controller $NAME status | grep 'ioFogController') ]]
}

function checkLegacyAgent() {
  local NS_CHECK=${2:-$NS}
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" legacy agent $1 status | grep 'RUNNING') ]]
  [[ "ok" == $(iofogctl -v -n "$NS_CHECK" legacy agent $1 status | grep 'Connection to Controller' | awk '{print $5}') ]]
}

function checkMovedMicroservice() {
  local MSVC="$1"
  local NEW_AGENT="$2"
  [[ ! -z $(iofogctl -v get microservices | grep $MSVC | grep $NEW_AGENT) ]]
}

function checkRenamedResource() {
  local RSRC=$1
  local OLDNAME=$2
  local NEWNAME=$3
  local NAMESPACE=$4
  [[ -z $(iofogctl -n ${NAMESPACE} -v get ${RSRC} | grep -w ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -n ${NAMESPACE} -v get ${RSRC} | grep -w ${NEWNAME}) ]]
}

function checkRenamedApplication() {
  local OLDNAME=$1
  local NEWNAME=$2
  local NAMESPACE=$3

  [[ -z $(iofogctl -n ${NAMESPACE} -v get applications | awk '{print $1}' | grep ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -n ${NAMESPACE} -v get applications |  awk '{print $1}' | grep ${NEWNAME}) ]]
}

function checkNamespaceExistsNegative() {
  local CHECK_NS="$1"
  [ -z "$(iofogctl get namespaces | grep $CHECK_NS)" ]
}

function checkRenamedNamespace() {
  local OLDNAME=$1
  local NEWNAME=$2
  [[ -z $(iofogctl -v get namespaces | grep -w ${OLDNAME}) ]]
  [[ ! -z $(iofogctl -v get namespaces | grep -w ${NEWNAME}) ]]
}

function hitMsvcEndpoint() {
  local PUBLIC_ENDPOINT="$1"
  local ITER=0
  local COUNT=0
  while [ $COUNT -eq 0 ] && [ $ITER -lt 24 ]; do
    sleep 10
    run curlMsvc "$PUBLIC_ENDPOINT"
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
  local USER=$1
  local HOST=$2
  local PORT=$3
  local KEY_FILE=$4
  local RESOURCE=$5

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

function checkRoute() {
  local NAME="$1"
  local FROM="$2"
  local TO="$3"

  local ALL=$(iofogctl -v -n "$NS" get all)
  echo "$ALL"
  echo "$ALL" | grep "ROUTE"
  echo "$ALL" | grep "$NAME"

  local GET=$(iofogctl -v -n "$NS" get routes)
  echo "$GET"
  echo "$GET" | grep "ROUTE"
  echo "$GET" | grep "$NAME"

  local DESC=$(iofogctl -v -n "$NS" describe route $APPLICATION_NAME/"$NAME")
  echo "$DESC"
  echo "$DESC" | grep "name: $NAME"
  echo "$DESC" | grep "from: $FROM"
  echo "$DESC" | grep "to: $TO"
}

function checkRouteNegative() {
  local NAME="$1"
  local FROM="$2"
  local TO="$3"

  local ALL=$(iofogctl -v -n "$NS" get all)
  echo "$ALL"
  [ -z "$(echo $ALL | grep $NAME)" ]

  local GET=$(iofogctl -v -n "$NS" get routes)
  echo "$GET"
  [ -z "$(echo $GET | grep $NAME)" ]
}