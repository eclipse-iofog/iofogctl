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

@test "Agent legacy commands" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Get Agent logs" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
  done
}

@test "Prune Agent" {
  initAgents
  local AGENT_NAME="${NAME}-0"
  iofogctl -v -n "$NS" prune agent "$AGENT_NAME"
  local CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  # TODO: Enable check that is not flake
  #checkAgentPruneController "$CONTROLLER_ENDPOINT" "$SSH_KEY_PATH"
}

@test "Disconnect from cluster" {
  initAgents
  iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkAgentsNegative
}

@test "Connect to cluster using deploy file" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect -f test/conf/k8s.yaml
  checkController
  checkAgents
}

@test "Disconnect from cluster again" {
  initAgents
  iofogctl -v -n "$NS" disconnect
  checkControllerNegative
  checkAgentsNegative
}

@test "Connect to cluster using flags" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect --name "$NAME" --kube "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController
  checkAgents
}

@test "Set default namespace" {
  iofogctl -v configure default-namespace "$NS"
}

@test "Deploy application" {
  initApplicationFiles
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc func-app-server "$NS"
  waitForMsvc func-app-ui "$NS"
}

@test "Test Public Ports w/ Microservices on same Agent" {
  initAgents
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  waitForProxyMsvc ${HOSTS[0]} ${USERS[0]} $SSH_KEY_PATH
  # Wait for public port to be up
  sleep 60
  # Wait for k8s service
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  # Hit the endpoint
  COUNT=$(curl -s --max-time 120 http://${EXT_IP}:5000/api/raw | jq '. | length')
  [ $COUNT -gt 0 ]
}

@test "Move microservice to another agent" {
  iofogctl -v move microservice $MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
  initAgents
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  waitForProxyMsvc ${HOSTS[1]} ${USERS[1]} $SSH_KEY_PATH
  waitForMsvc "$MSVC2_NAME" "$NS"
}

@test "Test Public Ports w/ Microservice on different Agents" {
  # Wait for public port to be up
  sleep 60
  # Wait for k8s service
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  # Hit the endpoint
  COUNT=$(curl -s --max-time 120 http://${EXT_IP}:5000/api/raw | jq '. | length')
  [ $COUNT -gt 0 ]
}

@test "Change Microservice Ports" {
  initApplicationFiles
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  # Change port
  sed -i.bak  "s/public: 5000/public: 6000/g" test/conf/application.yaml
  iofogctl -v deploy -f test/conf/application.yaml
  # Wait for port to update to 6000
  PORT=""
  SECS=0
  MAX=60
  while [[ $SECS -lt $MAX && ! -z $PORT ]]; do
    PORT=$(kubectl describe svc http-proxy -n "$NS" | grep 6000/TCP)
    SECS=$((SECS+1))
    sleep 1
  done
  # Check what happened in the loop above
  [ $SECS -lt $MAX ]
  waitForSvc "$NS" http-proxy
}

@test "Delete Public Port" {
  initApplicationFiles
  # Remove port info from the file
  sed -i.bak "s/.*ports:.*//g" test/conf/application.yaml
  sed -i.bak "s/.*external:.*//g" test/conf/application.yaml
  sed -i.bak "s/.*internal:.*//g" test/conf/application.yaml
  sed -i.bak "s/.*public:.*//g" test/conf/application.yaml

  # Update application
  iofogctl -v deploy -f test/conf/application.yaml

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

# Delete all does not delete application
@test "Delete application" {
  iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Deploy Agents for idempotence" {
  initAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Configure Controller" {
  for resource in controller; do
    iofogctl -v -n "$NS" configure "$resource" "$NAME" --kube "$KUBE_CONFIG"
  done
  iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Configure Agents" {
  initAgents
  iofogctl -v -n "$NS" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Detach agent" {
  local AGENT_NAME="${NAME}-0"
  iofogctl -v detach agent "$AGENT_NAME"
  checkAgentNegative "$AGENT_NAME"
  checkDetachedAgent "$AGENT_NAME"
}

@test "Attach agent" {
  local AGENT_NAME="${NAME}-0"
  iofogctl -v attach agent "$AGENT_NAME"
  checkAgent "$AGENT_NAME"
  checkDetachedAgentNegative "$AGENT_NAME"
}

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
        proxy: $PROXY_IMAGE
        kubelet: $KUBELET_IMAGE" > test/conf/k8s.yaml

  iofogctl -v -n "$NS" deploy -f test/conf/k8s.yaml
  checkController
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
