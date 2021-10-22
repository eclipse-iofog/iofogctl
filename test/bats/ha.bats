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
# ROUTER_IMAGE
# PROXY_IMAGE
# OPERATOR_IMAGE
# KUBELET_IMAGE
# DB_PROVIDER
# DB_USER
# DB_HOST
# DB_PORT
# DB_PW
# DB_NAME

. test/func/include.bash

NS="$NAMESPACE"

@test "Initialize tests" {
  stopTest
}

@test "Verify Agents >= 2" {
  startTest
  testAgentCount
  stopTest
}

@test "Verify kubectl works" {
  startTest
  kctl get ns
  stopTest
}

@test "Create namespace" {
  startTest
  iofogctl create namespace "$NS"
  stopTest
}

@test "Deploy Control Plane" {
  startTest
  echo "---
apiVersion: iofog.org/v3
kind: KubernetesControlPlane
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
  config: $KUBE_CONFIG
  replicas:
    controller: 2
  images:
    controller: $CONTROLLER_IMAGE
    operator: $OPERATOR_IMAGE
    portManager: $PORT_MANAGER_IMAGE
    proxy: $PROXY_IMAGE
    router: $ROUTER_IMAGE
    kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkControllerK8s
  checkControllerK8s
  stopTest
}

@test "Get endpoint" {
  startTest
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | head -n 1 | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
  stopTest
}

@test "Deploy Agents" {
  startTest
  initRemoteAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
  stopTest
}

@test "List Agents multiple times" {
  startTest
  initAgents
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  login "$CONTROLLER_ENDPOINT" "$USER_EMAIL" "$USER_PW"
  for IDX in $(seq 1 3); do
    checkAgentListFromController
  done
  stopTest
}

# TODO: Enable when no longer get connection refused when scaling replicas down
#@test "Delete Controller Instances and List Agents multiple times" {
#  startTest
#  initAgents
#  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
#  for REPLICAS in $(seq 3 4); do
#    kctl scale deployment controller --replicas $REPLICAS -n $NS
#    kctl rollout status deployment controller -n $NS
#    checkAgentListFromController
#  done
#  kctl scale deployment controller --replicas 2 -n $NS
#  kctl rollout status deployment controller -n $NS
#  stopTest
#}

@test "Deploy Agents again" {
  startTest
  initRemoteAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
  # Wait for router microservice
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    # Wait for router microservice
    waitForSystemMsvc "router" ${HOSTS[IDX]} ${USERS[IDX]} $SSH_KEY_PATH 
  done
  stopTest
}

# LOAD: test/bats/common-k8s.bats

@test "Delete Agents" {
  startTest
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" delete agent "$AGENT_NAME"
  done
  checkAgentsNegative
  stopTest
}

@test "Delete all" {
  startTest
  iofogctl -v -n "$NS" delete all
  checkControllerNegativeK8s
  checkControllerNegativeK8s
  checkAgentsNegative
  stopTest
}

@test "Delete namespace" {
  startTest
  iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
  stopTest
}
