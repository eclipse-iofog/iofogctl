#!/usr/bin/env bash

# Required environment variables
# NAMESPACE
# KUBE_CONFIG
# AGENT_LIST
# VANILLA_CONTROLLER
# KEY_FILE
# AGENT_PACKAGE_CLOUD_TOKEN
# CONTROLLER_IMAGE
# PORT_MANAGER_IMAGE
# PROXY_IMAGE
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

@test "Verify kubectl works" {
  kctl get ns
}

@test "Create namespace" {
  iofogctl create namespace "$NS"
}

@test "Deploy Control Plane" {
  echo "---
apiVersion: iofog.org/v1
kind: ControlPlane
metadata:
  name: ha-controlplane
spec:
  database:
    provider: $DB_PROVIDER
    user: $DB_USER
    host: $DB_HOST
    port: $DB_PORT
    password: $DB_PW
    databaseName: $DB_NAME
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
      replicas: 2
      images:
        operator: $OPERATOR_IMAGE
        portManager: $PORT_MANAGER_IMAGE
        proxy: $PROXY_IMAGE
        kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
}

@test "Get endpoint" {
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
}

@test "Deploy Agents" {
  initAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
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
  local CTRL_LIST=$(kctl get pods -l name=controller -n "$NS" | tail -n +2 | awk '{print $1}')
  local SAFE_CTRL=$(echo "$CTRL_LIST" | tail -n 1)
  for IDX in 0 1 2 3 4; do
    CTRL_LIST=$(kctl get pods -l name=controller -n "$NS" | tail -n +2 | awk '{print $1}')
    while read -r line; do
      if [ "$line" != "$SAFE_CTRL" ]; then
        kctl delete pods/"$line" -n "$NS" &
      fi
    done <<< "$CTRL_LIST"
    checkAgentListFromController
  done
}

@test "Deploy Agents again" {
  initAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Delete Agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" delete agent "$AGENT_NAME"
  done
  checkAgentsNegative
}

@test "Delete all" {
  iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkAgentsNegative
}

@test "Delete namespace" {
  iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
