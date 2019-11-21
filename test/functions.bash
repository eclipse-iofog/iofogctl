#!/usr/bin/env bash

function test(){
    eval "$@"
    [[ $? == 0 ]]
}

function initVanillaController(){
  VANILLA_USER=$(echo "$VANILLA_CONTROLLER" | sed "s|@.*||g")
  VANILLA_HOST=$(echo "$VANILLA_CONTROLLER" | sed "s|.*@||g")
  VANILLA_PORT=$(echo "$VANILLA_CONTROLLER" | cut -d':' -s -f2)
  VANILLA_PORT="${PORT:-22}"
}

function initAllLocalDeleteFile() {
  cat test/conf/local.yaml > test/conf/all-local.yaml
  echo "" >> test/conf/all-local.yaml
  cat test/conf/local-agent.yaml >> test/conf/all-local.yaml
  echo "" >> test/conf/all-local.yaml
  cat test/conf/application.yaml >> test/conf/all-local.yaml
}

function initMicroserviceFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Microservice 
metadata:
  name: ${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
    config:
      memoryLimit: 8192
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  rootHostAccess: false
  application: ${APPLICATION_NAME}
  volumes:
    - hostDestination: /tmp/microservice
      containerDestination: /tmp
      accessMode: rw
  ports:
    - internal: 443
      external: 5005
  env:
    - key: TEST
      value: 42
  routes:
    - ${MSVC1_NAME}
    - ${MSVC2_NAME}
  config:
    test_mode: true
    data_label: 'Anonymous_Person_2'" > test/conf/microservice.yaml
}

function initMicroserviceUpdateFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Microservice
metadata:
  name: ${MICROSERVICE_NAME}
spec:
  agent:
    name: ${NAME}-0
    config:
      memoryLimit: 5555
      diskDirectory: /tmp/iofog-agent/
  images:
    arm: edgeworx/healthcare-heart-rate:test-arm
    x86: edgeworx/healthcare-heart-rate:test
    registry: remote # public docker
  rootHostAccess: false
  application: ${APPLICATION_NAME}
  volumes:
    - hostDestination: /tmp/updatedmicroservice
      containerDestination: /tmp
      accessMode: rw
  ports:
    - internal: 443
      external: 5443
    - internal: 80
      external: 5080
  env:
    - key: TEST
      value: 75
    - key: TEST_2
      value: 42
  routes:
    - ${MSVC1_NAME}
  config:
    test_mode: true
    test_data:
      key: 42
    data_label: 'Anonymous_Person_3'" > test/conf/updatedMicroservice.yaml
}

function initApplicationFiles() {
  MSVCS="
    microservices:
    - name: $MSVC1_NAME
      agent:
        name: ${NAME}-0
        config:
          bluetoothEnabled: true # this will install the iofog/restblue microservice
          abstractedHardwareEnabled: false
      images:
        arm: edgeworx/healthcare-heart-rate:arm-v1
        x86: edgeworx/healthcare-heart-rate:x86-v1
        registry: remote # public docker
      rootHostAccess: false
      volumes:
        - hostDestination: /tmp/msvc
          containerDestination: /tmp
          accessMode: z
      ports: []
      config:
        test_mode: true
        data_label: 'Anonymous_Person'
    # Simple JSON viewer for the heart rate output
    - name: $MSVC2_NAME
      agent:
        name: ${NAME}-0
      images:
        arm: edgeworx/healthcare-heart-rate-ui:arm
        x86: edgeworx/healthcare-heart-rate-ui:x86
        registry: remote
      rootHostAccess: false
      ports:
        # The ui will be listening on port 80 (internal).
        - external: 5000 # You will be able to access the ui on <AGENT_IP>:5000
          internal: 80 # The ui is listening on port 80. Do not edit this.
          publicMode: false # Do not edit this.
      volumes: []
      env:
        - key: BASE_URL
          value: http://localhost:8080/data"
  ROUTES="
    routes:
    # Use this section to configure route between microservices
    # Use microservice name
    - from: $MSVC1_NAME
      to: $MSVC2_NAME"

  echo -n "---
  apiVersion: iofog.org/v1
  kind: Application
  metadata:
    name: $APPLICATION_NAME
  spec:" > test/conf/application.yaml
  echo -n "$MSVCS" >> test/conf/application.yaml
  echo "$ROUTES" >> test/conf/application.yaml
}

function initLocalAgentFile() {
  echo "---
apiVersion: iofog.org/v1
kind: Agent
metadata:
  name: ${NAME}-0
spec:
  host: 127.0.0.1
  container:
    image: ${AGENT_IMAGE}" > test/conf/local-agent.yaml
}

function initLocalControllerFile() {
    echo "---
apiVersion: iofog.org/v1
kind: ControlPlane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV
  controllers:
  - name: $NAME
    host: 127.0.0.1
    container:
      image: ${CONTROLLER_IMAGE}
---
apiVersion: iofog.org/v1
kind: Connector
metadata:
  name: $NAME
spec:
  host: localhost
  container:
    image: ${CONNECTOR_IMAGE}" > test/conf/local.yaml
}

function initAgentsFile() {
  initAgents
  # Empty file
  echo -n "" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    echo "---
apiVersion: iofog.org/v1
kind: Agent 
metadata:
  name: $AGENT_NAME
spec:
  host: ${HOSTS[$IDX]}
  ssh:
    user: ${USERS[$IDX]}
    keyFile: $KEY_FILE" >> test/conf/agents.yaml

  echo "====> Agent File:"
  cat test/conf/agents.yaml
  done
}

function initAgents(){
  USERS=()
  HOSTS=()
  PORTS=()
  AGENT_NAMES=()
  AGENTS=($AGENT_LIST)
  for AGENT in "${AGENTS[@]}"; do
    local USER=$(echo "$AGENT" | sed "s|@.*||g")
    local HOST=$(echo "$AGENT" | sed "s|.*@||g")
    local PORT=$(echo "$AGENT" | cut -d':' -s -f2)
    local PORT="${PORT:-22}"

    USERS+=" "
    USERS+="$USER"
    HOSTS+=" "
    HOSTS+="$HOST"
    PORTS+=" "
    PORTS+="$PORT"
    AGENT_NAMES+=" "
    AGENT_NAMES+="$AGENT_NAME"
  done
  USERS=($USERS)
  HOSTS=($HOSTS)
  PORTS=($PORTS)
}

function checkController() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe controller "$NAME" | grep "name: $NAME") ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe controlplane | grep "name: $NAME") ]]
}

function checkConnector() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" == $(iofogctl -v -n "$NS_CHECK" get connectors | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe connector "$NAME" | grep "name: $NAME") ]]
}

function checkConnectors() {
  NS_CHECK=$NS
  for CNCT in "$@"; do
    [[ "$CNCT" == $(iofogctl -v -n "$NS_CHECK" get connectors | grep "$CNCT" | awk '{print $1}') ]]
    [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe connector "$CNCT" | grep "name: $CNCT") ]]
  done
}

function checkControllerNegative() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkConnectorNegative() {
  NS_CHECK=${1:-$NS}
  [[ "$NAME" != $(iofogctl -v -n "$NS_CHECK" get connectors | grep "$NAME" | awk '{print $1}') ]]
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
  [[ "$APPLICATION_NAME" == $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  [[ "$MSVC1_NAME," == $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $3}') ]]
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS_CHECK" get applications | grep "$APPLICATION_NAME" | awk '{print $4}') ]]
  [[ "$MSVC1_NAME" == $(iofogctl -v -n "$NS_CHECK" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
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
  AGENT_NAME=$1
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS_CHECK" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" describe agent "$AGENT_NAME" | grep "name: $AGENT_NAME") ]]
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

function login() {
  local API_ENDPOINT="$1"
  local EMAIL="$2"
  local PASSWORD="$3"
  local LOGIN=$(curl --request POST \
--url $API_ENDPOINT/api/v3/user/login \
--header 'Content-Type: application/json' \
--data "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")
  echo $LOGIN
  ACCESS_TOKEN=$(echo $LOGIN | jq -r .accessToken)
  [[ ! -z "$ACCESS_TOKEN" ]]
  echo "$ACCESS_TOKEN" > /tmp/access_token.txt
  echo "$API_ENDPOINT" > /tmp/api_endpoint.txt
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

function checkLegacyConnector() {
  NS_CHECK=${1:-$NS}
  [[ ! -z $(iofogctl -v -n "$NS_CHECK" legacy connector $NAME status | grep 'is up and running') ]]
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