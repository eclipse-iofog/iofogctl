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
    connector: $CONNECTOR_IMAGE
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

@test "Deploy Connector" {
  echo "---
name: $NAME
replicas: 1
kubeconfig: $KUBE_CONFIG" > test/conf/cnct.yaml
  test iofogctl -v -n "$NS" deploy connector -f test/conf/cnct.yaml
  checkConnector
}

@test "Deploy Agents" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "List Agents multiple times" {
  initAgents
  login "$CONTROLLER_ENDPOINT" "$USER_EMAIL" "$USER_PW"
  for IDX in 0 1 2 3 4 5; do
    checkAgentListFromController
  done
}

@test "Delete Controller Instance and List Agents multiple times" {
  initAgents
  for IDX in 0 1 2 3 4 5; do
    kubectl delete pods -n "$NS" $(kubectl get pods -l name=connector -n "$NS" | awk 'FNR == 2 {print $1}') &
    checkAgentListFromController
  done
}

@test "Deploy Agents again" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
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