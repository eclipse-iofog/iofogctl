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
        kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
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
  checkAgentsNegative
}

@test "Connect to cluster using deploy file" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect -f test/conf/k8s.yaml
  checkController
  checkAgents
}

@test "Disconnect from cluster again" {
  initAgents
  test iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkAgentsNegative
}

@test "Connect to cluster using flags" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  test iofogctl -v -n "$NS" connect --name "$NAME" --kube "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
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

@test "Test Public Ports on Cluster" {
  # Wait for k8s service
  EXT_IP=$(waitForSvc http-proxy)
  # Hit the endpoint
  test curl http://${EXT_IP}:5000
}

@test "Change Microservice Ports" {
  initApplicationFiles
  EXT_IP=$(waitForSvc http-proxy)
  # Change port
  sed -i '' "s/external: 5000/external: 6000/g" test/conf/application.yaml
  test iofogctl -v deploy -f test/conf/application.yaml
  # Wait for port to update to 6000
  PORT=0
  SECS=0
  while [ $SECS -lt 30 && $PORT != 6000 ]; do
    PORT=$(kubectl describe svc http-proxy | grep 6000/TCP)
    SECS=$((SECS+1))
    sleep 1
  done
  # Check what happened in the loop above
  [ $SECS -lt 30 ]
  test kubectl describe svc http-proxy | grep 6000/TCP

  # Check service was not deleted
  NEW_IP=$(waitForSvc http-proxy)
  [[ $EXT_IP == $NEW_IP ]]
}

@test "Delete Public Port" {
  initApplicationFiles
  # Remove port info from the file
  sed -i '' "s/.*ports:.*//g" test/conf/application.yaml
  sed -i '' "s/.*external:.*//g" test/conf/application.yaml
  sed -i '' "s/.*internal:.*//g" test/conf/application.yaml
  sed -i '' "s/.*public:.*//g" test/conf/application.yaml

  # Update application
  test iofogctl -v deploy -f test/conf/application.yaml

  # Wait for port to be deleted
  EXIT_CODE=0
  SECS=0
  while [ $SECS -lt 30 && $EXIT_CODE -eq 0 ]; do
    EXIT_CODE=$(kubectl describe svc http-proxy | grep 6000/TCP)
    SECS=$((SECS+1))
    sleep 1
  done
  # Check what happened in the loop above
  [ $SECS -lt 30 ]
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

@test "Configure Controller" {
  for resource in controller; do
    test iofogctl -v -n "$NS" configure "$resource" "$NAME" --kube "$KUBE_CONFIG"
  done
  test iofogctl -v -n "$NS" logs controller "$NAME"
  checkLegacyController
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

@test "Deploy Controller for idempotence" {
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
        portManager: $PORT_MANAGER
        kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkAgentsNegative
}

@test "Delete namespace" {
  test iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
