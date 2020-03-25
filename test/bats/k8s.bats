#!/usr/bin/env bash

# Required environment variables
# NAMESPACE
# KUBE_CONFIG
# TEST_KUBE_CONFIG
# AGENT_LIST
# VANILLA_CONTROLLER
# KEY_FILE
# AGENT_PACKAGE_CLOUD_TOKEN
# CONTROLLER_IMAGE
# PORT_MANAGER_IMAGE
# PROXY_IMAGE
# ROUTER_IMAGE
# SCHEDULER_IMAGE
# OPERATOR_IMAGE
# KUBELET_IMAGE
# VANILLA_VERSION

. test/func/include.bash

NS="$NAMESPACE"

@test "Verify kubectl works" {
  kctl get ns
}

@test "Create namespace" {
  iofogctl create namespace "$NS"
}

@test "Deploy Control Plane" {
  echo "---
apiVersion: iofog.org/v2
kind: KubernetesControlPlane
metadata:
  name: func-controlplane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  config: $KUBE_CONFIG
  images:
    controller: $CONTROLLER_IMAGE
    operator: $OPERATOR_IMAGE
    portManager: $PORT_MANAGER_IMAGE
    proxy: $PROXY_IMAGE
    router: $ROUTER_IMAGE
    kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkControllerK8s "${K8S_POD}1"
}

@test "Get endpoint" {
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
}

@test "Get Controller logs on K8s after deploy" {
  iofogctl -v -n "$NS" logs controller "$NAME" | grep "api/v3"
}

@test "Deploy Agents" {
  initAgentsFile
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
} 

# LOAD: test/bats/common-k8s.bats

@test "Delete Agents" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" delete agent "$AGENT_NAME"
  done
  checkAgentsNegative
}

@test "Deploy Controller for idempotence" {
  echo "---
apiVersion: iofog.org/v2
kind: KubernetesControlPlane
metadata:
  name: func-controlplane
spec:
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
  config: $KUBE_CONFIG
  images:
    controller: $CONTROLLER_IMAGE
    operator: $OPERATOR_IMAGE
    portManager: $PORT_MANAGER
    proxy: $PROXY_IMAGE
    router: $ROUTER_IMAGE
    kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkControllerK8s "${K8S_POD}1"
}

@test "Delete all" {
  iofogctl -v -n "$NS" delete all
  checkControllerNegativeK8s "${K8S_POD}1"
  checkAgentsNegative
}

@test "Delete namespace" {
  iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
