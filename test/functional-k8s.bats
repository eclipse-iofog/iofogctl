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

NS=$(echo "$NAMESPACE""-k8s")

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy Controller" {
  echo "controlplane:
  controllers:
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

@test "Deploy Agents" {
  initAgentsFile
  test iofogctl -q -n "$NS" deploy -f test/conf/agents.yaml
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

@test "Deploy Agents for idempotence" {
  initAgentsFile
  test iofogctl -q -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Deploy Controller for idempotence" {
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

  test iofogctl -q -n "$NS" deploy controlplane -f test/conf/k8s.yaml
  checkController
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