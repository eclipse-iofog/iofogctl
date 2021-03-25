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

@test "Test no executors" {
  startTest
  testNoExecutors
  stopTest
}

@test "Deploy Control Plane" {
  startTest
  echo "---
apiVersion: iofog.org/v3
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
  checkControllerK8s
  stopTest
}

@test "Get endpoint" {
  startTest
  CONTROLLER_ENDPOINT=$(iofogctl -v -n "$NS" describe controlplane | grep endpoint | sed "s|.*endpoint: ||")
  [[ ! -z "$CONTROLLER_ENDPOINT" ]]
  echo "$CONTROLLER_ENDPOINT" > /tmp/endpoint.txt
  stopTest
}

@test "Get Controller logs on K8s after deploy" {
  startTest
  iofogctl -v -n "$NS" logs controller "$NAME" | grep "api/v3"
  stopTest
}

@test "Configure kube config file" {
  startTest
  local NEW_KUBE="/tmp/new-kubeconfig"
  local TEST_KUBE="$NEW_KUBE"
  if [[ ! -z $WSL_KEY_FILE ]]; then
    NEW_KUBE="C:\tmp\new-kubeconfig"
    TEST_KUBE='C:\\tmp\\new-kubeconfig'
  fi
  iofogctl -v -n "$NS" configure controlplane --kube "$NEW_KUBE"
  DESC=$(iofogctl -v -n "$NS" describe controlplane)
  echo $NEW_KUBE
  echo $DESC
  iofogctl -v -n "$NS" describe controlplane | grep $TEST_KUBE
  iofogctl -v -n "$NS" configure controlplane --kube "$KUBE_CONFIG"
  stopTest
}

@test "Deploy Agents" {
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

@test "Deploy Controller for idempotence" {
  startTest
  echo "---
apiVersion: iofog.org/v3
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
  checkControllerK8s
  stopTest
}

@test "Delete all" {
  startTest
  iofogctl -v -n "$NS" delete all
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
