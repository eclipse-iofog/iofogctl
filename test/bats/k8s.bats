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
# SCHEDULER_IMAGE
# OPERATOR_IMAGE
# KUBELET_IMAGE
# VANILLA_VERSION

. test/func/include.bash

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

@test "Get Controller logs on K8s after deploy" {
  iofogctl -v -n "$NS" logs controller "$NAME"
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
    waitForSystemMsvc "quay.io/interconnectedcloud/qdrouterd:latest" ${HOSTS[IDX]} ${USERS[IDX]} $SSH_KEY_PATH 
  done
}

@test "Deploy Volumes" {
  DIR="/tmp/iofogctl_tests"
  initAgents
  echo "---
apiVersion: iofog.org/v1
kind: Volume
spec:
  source: $DIR
  destination: $DIR
  permissions: 666
  agents:
  - $NAME-0
  - $NAME-1" > test/conf/volume.yaml

  run mkdir $DIR
  for IDX in 1 2 3; do
    echo "test$IDX" > "$DIR/test$IDX"
  done
  run mkdir $DIR/testdir
  for IDX in 1 2 3; do
    echo "test$IDX" > "$DIR/testdir/test$IDX"
  done
  iofogctl -v -n "$NS" deploy -f test/conf/volume.yaml

  # Check files
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  for IDX in "${!AGENTS[@]}"; do
    for FILE_IDX in 1 2 3; do
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat /tmp/iofogctl_tests/test$FILE_IDX | grep "test$FILE_IDX"
      ssh -oStrictHostKeyChecking=no -i "$SSH_KEY_PATH" "${USERS[IDX]}@${HOSTS[IDX]}" -- cat /tmp/iofogctl_tests/testdir/test$FILE_IDX | grep "test$FILE_IDX"
    done
  done
}