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
. test/functional.vars.bash

NS="$NAMESPACE"
USER_PW="S5gYVgLEZV"
USER_EMAIL="user@domain.com"

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy Control Plane and Connector" {
  echo "---
controlplane:
  iofoguser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  controllers:
  - name: $NAME
    kubeconfig: $KUBE_CONFIG
  images:
    controller: $CONTROLLER_IMAGE
    connector: $CONNECTOR_IMAGE
    scheduler: $SCHEDULER_IMAGE
    operator: $OPERATOR_IMAGE
    kubelet: $KUBELET_IMAGE
connectors:
- name: $NAME
  kubeconfig: $KUBE_CONFIG" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
  checkConnector
}

@test "Get endpoint" {
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
}

@test "Controller legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
}

@test "Get Controller logs on K8s after deploy" {
  test iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Deploy Agents" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Agent legacy commands" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" legacy agent "$AGENT_NAME" status
  done
}

@test "Get Agent logs" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
  done
}

@test "Disconnect from cluster" {
  initAgents
  test iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Connect to cluster using Controller IP" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect "$NAME" --controller "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
  checkConnector
  checkAgents
}

@test "Disconnect from cluster again" {
  initAgents
  test iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Connect to cluster using Kube Config" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect "$NAME" --kube-config "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
  checkConnector
  checkAgents
}

@test "Deploy Agents for idempotence" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Delete Agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" delete agent "$AGENT_NAME"
  done
  checkAgentsNegative
  sleep 30 # Sleep to make sure vKubelet resolves with K8s API Server before we delete all
}

@test "Deploy Controller and Connector for idempotence" {
  echo "---
controlplane:
  iofoguser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  controllers:
  - name: $NAME
    kubeconfig: $KUBE_CONFIG
  images:
    controller: $CONTROLLER_IMAGE
    connector: $CONNECTOR_IMAGE
    scheduler: $SCHEDULER_IMAGE
    operator: $OPERATOR_IMAGE
    kubelet: $KUBELET_IMAGE
connectors:
- name: $NAME
  kubeconfig: $KUBE_CONFIG" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
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