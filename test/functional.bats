#!/usr/bin/env bash

# Required environment variables
# NAMESPACE
# KUBE_CONFIG
# AGENT_LIST
# VANILLA_CONTROLLER
# KEY_FILE
# PACKAGE_CLOUD_TOKEN
# CONTROLLER_IMAGE
# CONNECTOR_IMAGE
# SCHEDULER_IMAGE
# OPERATOR_IMAGE
# KUBELET_IMAGE
# VANILLA_VERSION

. test/functions.bash

NS="$NAMESPACE"
NAME="func_test"
APPLICATION_NAME="func_app"
MSVC1_NAME="func_app_server"
MSVC2_NAME="func_app_ui"

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
    local AGENT_NAME="${NAME}_${IDX}"

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
  AGENT_NAMES=($AGENT_NAMES)
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

function checkController() {
  [[ "$NAME" == $(iofogctl -q -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -q -n "$NS" describe controller "$NAME" | grep "name: $NAME") ]]
}

function checkControllerNegative() {
  [[ "$NAME" != $(iofogctl -q -n "$NS" get controllers | grep "$NAME" | awk '{print $1}') ]]
}

function checkApplication() {
  [[ "$APPLICATION_NAME" == $(iofogctl -q -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ ! -z $(iofogctl -q -n "$NS" describe application "$APPLICATION_NAME" | grep "name: $APPLICATION_NAME") ]]
  [[ "$MSVC1_NAME," == $(iofogctl -q -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $3}') ]]
  [[ "$MSVC2_NAME" == $(iofogctl -q -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $4}') ]]
  [[ "$MSVC1_NAME" == $(iofogctl -q -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  # Check config
  [[ "{\"data_label\":\"Anonymous Person\",\"test_mode\":true}" == $(iofogctl -q -n "$NS" get microservices | grep "$MSVC1_NAME" | awk -F '\t' '{print $4}') ]]
  [[ "bluetoothenabled: true" == $(iofogctl -q -n "$NS" describe agent "${NAME}_0" | grep bluetooth ) ]]
  # Check route
  [[ "$MSVC2_NAME" == $(iofogctl -q -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $5}') ]]
  # Check ports
  msvcWithPorts=$(iofogctl -q -n "$NS" get microservices | grep "5000:80")
  [[ "$MSVC2_NAME" == $(echo "$msvcWithPorts" | awk '{print $1}') ]]
  # Check volumes
  msvcWithVolume=$(iofogctl -q -n "$NS" get microservices | grep "/tmp/msvc:/tmp")
  [[ "$MSVC1_NAME" == $(echo "$msvcWithVolume" | awk '{print $1}') ]]

  # Check describe
  # TODO: Use another testing framework to verify proper output of yaml file
  iofogctl -q -n "$NS" describe application "$APPLICATION_NAME" -o "test/conf/app_output.yaml"
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
  [[ "$NAME" != $(iofogctl -q -n "$NS" get applications | grep "$APPLICATION_NAME" | awk '{print $1}') ]]
  [[ "$MSVC1_NAME" != $(iofogctl -q -n "$NS" get microservices | grep "$MSVC1_NAME" | awk '{print $1}') ]]
  [[ "$MSVC2_NAME" != $(iofogctl -q -n "$NS" get microservices | grep "$MSVC2_NAME" | awk '{print $1}') ]]
}

function checkAgents() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_$(((IDX++)))"
    [[ "$AGENT_NAME" == $(iofogctl -q -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
    [[ ! -z $(iofogctl -q -n "$NS" describe agent "$AGENT_NAME" | grep "name: $AGENT_NAME") ]]
  done
}

function checkAgentsNegative() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_$(((IDX++)))"
    [[ "$AGENT_NAME" != $(iofogctl -q -n "$NS" get agents | grep "$AGENT_NAME" | awk '{print $1}') ]]
  done
}

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy controller" {
  test iofogctl -q -n "$NS" deploy controller "$NAME" --kube-config "$KUBE_CONFIG"
  checkController
}

@test "Get credentials" {
  export CONTROLLER_EMAIL=$(iofogctl -q -n "$NS" describe controller "$NAME" | grep email | sed "s|.*email: ||")
  export CONTROLLER_PASS=$(iofogctl -q -n "$NS" describe controller "$NAME" | grep password | sed "s|.*password: ||")
  export CONTROLLER_ENDPOINT=$(iofogctl -q -n "$NS" describe controller "$NAME" | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_EMAIL" ]]
  [[ ! -z "$CONTROLLER_PASS" ]]
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_EMAIL" > /tmp/email.txt
  echo "$CONTROLLER_PASS" > /tmp/pass.txt
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
}


@test "Controller legacy commands after deploy" {
  sleep 15 # Sleep to avoid SSH tunnel bug from K8s
  test iofogctl -q -n "$NS" legacy controller "$NAME" iofog list
}

@test "Get Controller logs on K8s after deploy" {
  test iofogctl -q -n "$NS" logs controller "$NAME"
}

@test "Deploy agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    test iofogctl -q -n "$NS" deploy agent "$AGENT_NAME" --user "${USERS[IDX]}" --host "${HOSTS[IDX]}" --key-file "$KEY_FILE" --port "${PORTS[IDX]}"
  done
  checkAgents
}

@test "Agent legacy commands" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    test iofogctl -q -n "$NS" legacy agent "$AGENT_NAME" status
  done
}

@test "Controller legacy commands after connect with Kube Config" {
  test iofogctl -q -n "$NS" legacy controller "$NAME" iofog list
}

@test "Get Controller logs on K8s after connect with Kube Config" {
  test iofogctl -q -n "$NS" logs controller "$NAME"
}

@test "Get Agent logs" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    test iofogctl -q -n "$NS" logs agent "$AGENT_NAME"
  done
}

@test "Disconnect from cluster" {
  initAgents
  test iofogctl -q -n "$NS" disconnect
  checkControllerNegative
  checkAgentsNegative
}

@test "Connect to cluster using Controller IP" {
  CONTROLLER_EMAIL=$(cat /tmp/email.txt)
  CONTROLLER_PASS=$(cat /tmp/pass.txt)
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -q -n "$NS" connect "$NAME" --controller "$CONTROLLER_ENDPOINT" --email "$CONTROLLER_EMAIL" --pass "$CONTROLLER_PASS"
  checkController
  checkAgents
}

@test "Disconnect from cluster again" {
  initAgents
  test iofogctl -q -n "$NS" disconnect
  checkControllerNegative
  checkAgentsNegative
}

@test "Connect to cluster using Kube Config" {
  CONTROLLER_EMAIL=$(cat /tmp/email.txt)
  CONTROLLER_PASS=$(cat /tmp/pass.txt)
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -q -n "$NS" connect "$NAME" --kube-config "$KUBE_CONFIG" --email "$CONTROLLER_EMAIL" --pass "$CONTROLLER_PASS"
  checkController
  checkAgents
}

# TODO: Enable these if ever possible to do with IP connect
#@test "Get Controller logs after connect with IP" {
#  test iofogctl -q -n "$NS" logs controller "$NAME"
#}
#@test "Get Controller logs on K8s after connect with IP" {
#  test iofogctl -q -n "$NS" logs controller "$NAME"
#}

@test "Delete Agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    test iofogctl -q -n "$NS" delete agent "$AGENT_NAME"
  done
}

@test "Delete Controller" {
  test iofogctl -q -n "$NS" delete controller "$NAME"
  checkAgentsNegative
  checkControllerNegative
}

@test "Deploy Controller from file" {
  echo "controllers:
- name: $NAME
  kubeconfig: $KUBE_CONFIG
  images:
    controller: $CONTROLLER_IMAGE
    connector: $CONNECTOR_IMAGE
    scheduler: $SCHEDULER_IMAGE
    operator: $OPERATOR_IMAGE
    kubelet: $KUBELET_IMAGE
  iofoguser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV" > test/conf/k8s.yaml

  test iofogctl -q -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
}

@test "Deploy Agents from file" {
  initAgents
  echo "agents:" > test/conf/agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}_${IDX}"
    echo "- name: $AGENT_NAME
  user: ${USERS[$IDX]}
  host: ${HOSTS[$IDX]}
  keyfile: $KEY_FILE" >> test/conf/agents.yaml
  done

  test iofogctl -q -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Test Agent deploy for idempotence" {
  test iofogctl -q -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Test Controller deploy for idempotence" {
  test iofogctl -q -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
}

@test "Delete all" {
  test iofogctl -q -n "$NS" delete all
  checkControllerNegative
  checkAgentsNegative
}

# TODO: Enable this when a release of Controller is usable here (version needs to be specified for dev package)
#@test "Deploy vanilla Controller" {
#  initVanillaController
#  test iofogctl -q -n "$NS" deploy controller "$NAME" --user "$VANILLA_USER" --host "$VANILLA_HOST" --key-file "$KEY_FILE" --port "$VANILLA_PORT"
#  checkController
#}

@test "Deploy vanilla Controller" {
  initVanillaController
  echo "controllers:
- name: $NAME
  user: $VANILLA_USER
  host: $VANILLA_HOST
  port: $VANILLA_PORT
  keyfile: $KEY_FILE
  version: $VANILLA_VERSION
  packagecloudtoken: $PACKAGE_CLOUD_TOKEN
  iofoguser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV" > test/conf/vanilla.yaml

  test iofogctl -q -n "$NS" deploy -f test/conf/vanilla.yaml
  checkController
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -q -n "$NS" legacy controller "$NAME" iofog list
}

@test "Get Controller logs after vanilla deploy" {
  test iofogctl -q -n "$NS" logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
  test iofogctl -q -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -q -n "$NS" deploy application -f test/conf/application.yaml
  checkApplication
}

@test "Delete application" {
  test iofogctl -q -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Deploy application from root file" {
  test iofogctl -q -n "$NS" deploy -f test/conf/root_application.yaml
  checkApplication
}

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -q -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -q -n "$NS" delete all
  checkControllerNegative
  checkAgentsNegative
}

@test "Delete namespace" {
  test iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}