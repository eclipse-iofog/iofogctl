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
# DB_PROVIDER
# DB_USER
# DB_HOST
# DB_PORT
# DB_PW
# DB_NAME

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"
USER_PW="S5gYVgLEZV"
USER_EMAIL="user@domain.com"

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy Control Plane" {
  echo "---
database:
  provider: $DB_PROVIDER
  user: $DB_USER
  host: $DB_HOST
  port: $DB_PORT
  password: $DB_PW
  databasename: $DB_NAME
iofoguser:
  name: Testing
  surname: Functional
  email: $USER_EMAIL
  password: $USER_PW
controllers:
- name: $NAME
  kubeconfig: $KUBE_CONFIG
  replicas: 2
images:
  controller: $CONTROLLER_IMAGE
  scheduler: $SCHEDULER_IMAGE
  operator: $OPERATOR_IMAGE
  kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy controlplane -f test/conf/k8s.yaml
  checkController
}

@test "Get endpoint" {
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
}

@test "Deploy Connectors" {
  local CNCT_A="connector-a"
  local CNCT_B="connector-b"
  echo "---
connectors:
- name: $CNCT_A
  image: $CONNECTOR_IMAGE
  kubeconfig: $KUBE_CONFIG
- name: $CNCT_B
  image: $CONNECTOR_IMAGE
  kubeconfig: $KUBE_CONFIG" > test/conf/cncts.yaml
  test iofogctl -v -n "$NS" deploy -f test/conf/cncts.yaml
  checkConnectors "$CNCT_A" "$CNCT_B"
}

@test "Deploy Agents" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "List Agents multiple times" {
  initAgents
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  login "$CONTROLLER_ENDPOINT" "$USER_EMAIL" "$USER_PW"
  for IDX in 0 1 2 3 4 5; do
    checkAgentListFromController
  done
}

@test "Delete Controller Instances and List Agents multiple times" {
  initAgents
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  local CTRL_LIST=$(kubectl get pods -l name=controller -n "$NS" | tail -n +2 | awk '{print $1}')
  local SAFE_CTRL=$(echo "$CTRL_LIST" | tail -n 1)
  for IDX in 0 1 2 3 4; do
    CTRL_LIST=$(kubectl get pods -l name=controller -n "$NS" | tail -n +2 | awk '{print $1}')
    while read -r line; do
      if [ "$line" != "$SAFE_CTRL" ]; then
        kubectl delete pods/"$line" -n "$NS" &
      fi
    done <<< "$CTRL_LIST"
    checkAgentListFromController
  done
}

@test "Deploy Agents again" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Delete Agents" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" delete agent "$AGENT_NAME"
  done
  checkAgentsNegative
  sleep 30 # Sleep to make sure vKubelet resolves with K8s API Server before we delete all
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Delete namespace" {
  test iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}