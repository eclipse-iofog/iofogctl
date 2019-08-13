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

function initApplicationFiles() {
  APP="name: $APPLICATION_NAME"
  MSVCS="microservices:
  - name: $MSVC1_NAME
    agent:
      name: ${NAME}_0
      config:
        bluetoothenabled: true # this will install the iofog/restblue microservice
        abstractedhardwareEnabled: false
    images:
      arm: edgeworx/healthcare-heart-rate:arm-v1
      x86: edgeworx/healthcare-heart-rate:x86-v1
      registry: 1 # public docker
    roothostaccess: false
    volumes:
      - hostdestination: /tmp/msvc
        containerdestination: /tmp
        accessmode: z
    ports: []
    config:
      test_mode: true
      data_label: 'Anonymous Person'
  # Simple JSON viewer for the heart rate output
  - name: $MSVC2_NAME
    agent:
      name: ${NAME}_0
    images:
      arm: edgeworx/healthcare-heart-rate-ui:arm
      x86: edgeworx/healthcare-heart-rate-ui:x86
      registry: 1
    roothostaccess: false
    ports:
      # The ui will be listening on port 80 (internal).
      - external: 5000 # You will be able to access the ui on <AGENT_IP>:5000
        internal: 80 # The ui is listening on port 80. Do not edit this.
        publicmode: false # Do not edit this.
    volumes: []
    env:
      - key: BASE_URL
        value: http://localhost:8080/data"
  ROUTES="routes:
  # Use this section to configure route between microservices
  # Use microservice name
  - from: $MSVC1_NAME
    to: $MSVC2_NAME"

  echo "$APP" > test/conf/application.yaml
  echo "$MSVCS" >> test/conf/application.yaml
  echo "$ROUTES" >> test/conf/application.yaml
  echo -n "applications:
  - " > test/conf/root_application.yaml
  echo "$APP" >> test/conf/root_application.yaml
  echo "$MSVCS" | awk '{print "   ", $0}' >> test/conf/root_application.yaml
  echo "$ROUTES" | awk '{print "   ", $0}' >> test/conf/root_application.yaml
}

function initLocalAgentFile() {
  echo "agents:
    - name: ${NAME}_0
      host: 127.0.0.1" > test/conf/local-agent.yaml
}

function initAgentsFile() {
  initAgents
  echo "agents:" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    echo "- name: $AGENT_NAME
  user: ${USERS[$IDX]}
  host: ${HOSTS[$IDX]}
  keyfile: $KEY_FILE" > test/conf/agents.yaml
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
  [[ "$NAME" == $(iofogctl -v -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe controller "$NAME" | grep "name: $NAME") ]]
}

function checkControllerNegative() {
  [[ "$NAME" != $(iofogctl -v -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkApplication() {
  [[ "$APPLICATION_NAME" == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  [[ "$MSVC1_NAME," == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $3}') ]]
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $4}') ]]
  [[ "$MSVC1_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  # Check config
  [[ "{\"data_label\":\"Anonymous Person\",\"test_mode\":true}" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk -F '\t' '{print $4}') ]]
  [[ "bluetoothenabled: true" == $(iofogctl -v -n "$NS" describe agent "${NAME}_0" | grep bluetooth ) ]]
  # Check route
  [[ "$MSVC2_NAME" == $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $5}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -v -n "$NS" get microservices | grep "5000:80")
  [[ "$MSVC2_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -v -n "$NS" get microservices | grep "/tmp/msvc:/tmp")
  [[ "$MSVC1_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -v -n "$NS" describe application "$APPLICATION_NAME" -o "test/conf/app_output.yaml"
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $APPLICATION_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "name: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "routes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "from: $MSVC1_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "to: $MSVC2_NAME") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "ports:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "external: 5000") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "internal: 80") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "volumes:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "hostdestination: /tmp/msvc") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "containerdestination: /tmp") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "images:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "x86: edgeworx/healthcare-heart-rate:x86-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "arm: edgeworx/healthcare-heart-rate:arm-v1") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "env:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "key: BASE_URL") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "value: http://localhost:8080/data") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "config:") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "test_mode: true") ]]
  [[ ! -z $(cat test/conf/app_output.yaml | grep "data_label: Anonymous Person") ]]
  rm -f test/conf/app_output.yaml
}

function checkApplicationNegative() {
  [[ "$NAME" != $(iofogctl -v -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ "$MSVC1_NAME" != $(iofogctl -v -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  [[ "$MSVC2_NAME" != $(iofogctl -v -n "$NS" get microservices | grep "$MSVC2_NAME" | awk '{print $1}') ]]
}

function checkAgent() {
  AGENT_NAME=$1
  [[ "$AGENT_NAME" == $(iofogctl -v -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -v -n "$NS" describe agent "$AGENT_NAME" | grep "name: $AGENT_NAME") ]]
}

function checkAgentNegative() {
  AGENT_NAME=$1
  [[ "$AGENT_NAME" != $(iofogctl -v -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
}

function checkAgents() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_$(((IDX++)))"
    checkAgent "$AGENT_NAME"
  done
}

function checkAgentsNegative() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_$(((IDX++)))"
    checkAgentNegative "$AGENT_NAME"
  done
}