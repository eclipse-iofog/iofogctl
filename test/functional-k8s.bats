#!/usr/bin/env bash

# Required environment variables
# NAMESPACE
# KUBE_CONFIG
# AGENT_LIST
# VANILLA_CONTROLLER
# KEY_FILE
# AGENT_PACKAGE_CLOUD_TOKEN
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
apiVersion: iofog.org/v1
kind: ControlPlane
metadata:
  name: func-controlplane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  controllers:
  - name: $NAME
    container:
      image: $CONTROLLER_IMAGE
    kube:
      config: $KUBE_CONFIG
      images:
        operator: $OPERATOR_IMAGE
        kubelet: $KUBELET_IMAGE
---
apiVersion: iofog.org/v1
kind: Connector
metadata:
  name: $NAME
spec:
  container:
    image: $CONNECTOR_IMAGE
  kube:
    config: $KUBE_CONFIG" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
  checkConnector
}

@test "Get endpoint" {
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
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
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Get Agent logs" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
  done
}

@test "Prune Agent" {
  initAgents
  local AGENT_NAME="${NAME}-0"
  test iofogctl -v -n "$NS" prune agent "$AGENT_NAME"
  local CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  checkAgentPruneController "$CONTROLLER_ENDPOINT" "$SSH_KEY_PATH"
}

@test "Disconnect from cluster" {
  initAgents
  test iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Connect to cluster using deploy file" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect -f test/conf/k8s.yaml
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

@test "Connect to cluster using flags" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect --name "$NAME" --kube "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
  checkConnector
  checkAgents
}

@test "Set default namespace" {
  test iofogctl -v configure default-namespace "$NS"
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
}

@test "Move microservice to another agent" {
  test iofogctl -v move microservice $MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
}

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Deploy Agents for idempotence" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Configure Controller and Connector" {
  for resource in controller connector; do
    test iofogctl -v -n "$NS" configure "$resource" "$NAME" --kube "$KUBE_CONFIG"
  done
  test iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Configure Agents" {
  initAgents
  test iofogctl -v -n "$NS" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Detach agent" {
  local AGENT_NAME="${NAME}-0"
  test iofogctl -v detach agent "$AGENT_NAME"
  checkAgentNegative "$AGENT_NAME"
  checkDetachedAgent "$AGENT_NAME"
}

@test "Attach agent" {
  local AGENT_NAME="${NAME}-0"
  test iofogctl -v attach agent "$AGENT_NAME"
  checkAgent "$AGENT_NAME"
  checkDetachedAgentNegative "$AGENT_NAME"
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
apiVersion: iofog.org/v1
kind: ControlPlane
metadata:
  name: func-controlplane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  controllers:
  - name: $NAME
    container:
      image: $CONTROLLER_IMAGE
    kube:
      config: $KUBE_CONFIG
      images:
        operator: $OPERATOR_IMAGE
        kubelet: $KUBELET_IMAGE
---
apiVersion: iofog.org/v1
kind: Connector 
metadata:
  name: $NAME
spec:
  container:
    image: $CONNECTOR_IMAGE
  kube:
    config: $KUBE_CONFIG" > test/conf/k8s.yaml

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