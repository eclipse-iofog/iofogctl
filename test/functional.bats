#!/usr/bin/env bash

# Required environment variables
# NAMESPACE
# KUBE_CONFIG
# AGENTS
# AGENTS_KEY_FILE
# CONTROLLER_IMAGE
# CONNECTOR_IMAGE
# SCHEDULER_IMAGE
# OPERATOR_IMAGE
# KUBELET_IMAGE

. ./functions.bash

NAME="func_test"

function initAgents(){
  USERS=()
  HOSTS=()
  PORTS=()
  for AGENT in "${AGENTS[@]}"; do
    local USER_HOST="${AGENT%:*}"
    local USER=$(sed "s|@.*||g")
    local HOST=$(sed "s|.*@||g")
    local PORT=$(echo "$AGENT" | cut -d':' -s -f2)
    local PORT="${PORT:-22}"

    USERS+="$USER"
    HOSTS+="$HOST"
    PORTS+="$PORT"
  done
}

function checkController() {
  [[ *"$NAME"* == $(iofogctl get controllers) ]]
  [[ *"$NAME"* == $(iofogctl get all) ]]
  [[ *"$NAME"* == $(iofogctl describe controller $NAME) ]]
}

function checkAgents() {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    forIofogCTL "deploy agent $AGENT_NAME --user ${USERS[IDX]} --host ${HOSTS[IDX]} --key-file $AGENTS_KEY_FILE --port ${PORTS[IDX]}"
    [[ *"$AGENT_NAME"* == $(iofogctl get agents) ]]
    [[ *"$AGENT_NAME"* == $(iofogctl get all) ]]
    [[ *"$AGENT_NAME"* == $(iofogctl describe agent $NAME) ]]
  done
}

@test "Deploy controller" {
  forIofogCTL "deploy controller $NAME --kube-config $KUBE_CONFIG"
  checkController
}

@test "Get credentials" {
  CONTROLLER_EMAIL=$(iofogctl describe controller "$NAME" | grep email | sed "|.*email: ||g")
  CONTROLLER_PASS=$(iofogctl describe controller "$NAME" | grep password | sed "|.*password: ||g")
  CONTROLLER_ENDPOINT=$(iofogctl describe controller "$NAME" | grep endpoint | sed "|.*endpoint: ||g")
  [[ ! -z "$CONTROLLER_EMAIL" ]]
  [[ ! -z "$CONTROLLER_PASS" ]]
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
}

@test "Deploy agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    forIofogCTL "deploy agent $AGENT_NAME --user ${USERS[IDX]} --host ${HOSTS[IDX]} --key-file $AGENTS_KEY_FILE --port ${PORTS[IDX]}"
  done
  checkAgents
}

@test "Disconnect from cluster" {
  forIofogCTL "disconnect"
  [[ *"$NAME"* != $(iofogctl get controllers) ]]
  [[ *"$NAME"* != $(iofogctl get all) ]]
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    [[ *"$AGENT_NAME"* != $(iofogctl get agents) ]]
    [[ *"$AGENT_NAME"* != $(iofogctl get all) ]]
  done
}

@test "Connect to cluster using Kube Config" {
  forIofogCTL "connect $NAME --kube-config $KUBE_CONFIG --email $CONTROLLER_EMAIL --pass $CONTROLLER_PASS"
  checkController
  checkAgents
}

@test "Connect to cluster using Controller IP" {
  forIofogCTL "disconnect"
  forIofogCTL "connect $NAME --controller $CONTROLLER_ENDPOINT --email $CONTROLLER_EMAIL --pass $CONTROLLER_PASS"
  checkController
  checkAgents
}

@test "Delete Agents" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    forIofogCTL "delete agent $AGENT_NAME"
    [[ *"$AGENT_NAME"* != $(iofogctl get agents) ]]
    [[ *"$AGENT_NAME"* != $(iofogctl get all) ]]
    [[ *"$AGENT_NAME"* != $(iofogctl describe agent "$NAME") ]]
  done
}

@test "Delete Controller" {
  forIofogCTL "delete controller $NAME"
  [[ *"$NAME"* != $(iofogctl get controllers) ]]
  [[ *"$NAME"* != $(iofogctl get all) ]]
  [[ *"$NAME"* != $(iofogctl describe controller $NAME) ]]
}

@test "Deploy Controller and Agents from file" {
  sed -i.bak "s|<NAME>|$NAME|g" ./controller.yaml
  sed -i.bak "s|<KUBE_CONFIG>|$KUBE_CONFIG|g" ./controller.yaml
  sed -i.bak "s|<CONTROLLER_IMAGE>|$CONTROLLER_IMAGE|g" ./controller.yaml
  sed -i.bak "s|<CONNECTOR_IMAGE>|$CONNECTOR_IMAGE|g" ./controller.yaml
  sed -i.bak "s|<SCHEDULER_IMAGE>|$SCHEDULER_IMAGE|g" ./controller.yaml
  sed -i.bak "s|<OPERATOR_IMAGE>|$OPERATOR_IMAGE|g" ./controller.yaml
  sed -i.bak "s|<KUBELET_IMAGE>|$KUBELET_IMAGE|g" ./controller.yaml

  echo "agents:" > ./agents.yaml
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    echo " - name: $AGENT_NAME" >> ./agents.yaml
    echo "   user: ${USERS[IDX]}" >> ./agents.yaml
    echo "   host: ${HOSTS[IDX}" >> ./agents.yaml
    echo "   keyfile: $AGENTS_KEY_FILE" >> ./agents.yaml
  done

  forIofogCTL "deploy -f ./controller.yaml"
  checkController

  forIofogCTL "deploy -f ./agents.yaml"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    forIofogCTL "deploy agent $AGENT_NAME --user ${USERS[IDX]} --host ${HOSTS[IDX]} --key-file $AGENTS_KEY_FILE --port ${PORTS[IDX]}"
  done
  checkAgents
}

@test "Delete all" {
  forIofogCTL "delete all"
  [[ *"$NAME"* != $(iofogctl get controllers) ]]
  [[ *"$NAME"* != $(iofogctl get all) ]]
  [[ *"$NAME"* != $(iofogctl describe controller $NAME) ]]
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="$NAME_$(((IDX++)))"
    forIofogCTL "deploy agent $AGENT_NAME --user ${USERS[IDX]} --host ${HOSTS[IDX]} --key-file $AGENTS_KEY_FILE --port ${PORTS[IDX]}"
    [[ *"$AGENT_NAME"* != $(iofogctl get agents) ]]
    [[ *"$AGENT_NAME"* != $(iofogctl get all) ]]
    [[ *"$AGENT_NAME"* != $(iofogctl describe agent $NAME) ]]
  done
}