@test "Deploy Volumes" {
  testDeployVolume
}

@test "Get and Describe Volumes" {
  testGetDescribeVolume
}

@test "Delete Volumes and Redeploy" {
  testDeleteVolume
  testDeployVolume
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
  checkNamespaceExistsNegative "$NS"
}

@test "Connect to cluster using deploy file" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect -f test/conf/k8s.yaml
  checkControllerK8s "${K8S_POD}1"
  checkAgents
}

@test "Disconnect from cluster again" {
  initAgents
  iofogctl -v -n "$NS" disconnect
  checkNamespaceExistsNegative "$NS"
}

@test "Connect to cluster using flags" {
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect --kube "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkControllerK8s "${K8S_POD}1"
  checkAgents
}

@test "Set default namespace" {
  iofogctl -v configure default-namespace "$NS"
}

@test "Deploy application" {
  initApplicationFiles
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
}

@test "Volumes are mounted" {
  testMountVolume
}

@test "Test Public Ports w/ Microservices on same Agent" {
  initAgents
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  waitForProxyMsvc ${HOSTS[0]} ${USERS[0]} $SSH_KEY_PATH
  # Wait for k8s service
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  testDefaultProxyConfig "$EXT_IP"
  # Hit the endpoint
  hitMsvcEndpoint "$EXT_IP"
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
  # Wait for k8s service
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  testDefaultProxyConfig "$EXT_IP"
  hitMsvcEndpoint "$EXT_IP"
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
    PORT=$(kctl describe svc http-proxy -n "$NS" | grep 6000/TCP)
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
  RET=0
  SECS=0
  while [ $SECS -lt 30 ] && [ -z "$RET" ]; do
    RET=$(kctl describe svc http-proxy | grep 6000/TCP)
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
  initRemoteAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
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

@test "Detach with same name" {
  local A0="${NAME}-0"
  local A1="${NAME}-1"
  # Rename and fail
  iofogctl -v rename agent $A1 $A0
  run iofogctl -v detach agent $A0
  [ "$status" -eq 1 ]
  # Rename attached and succeed
  iofogctl -v rename agent $A0 $A1
  iofogctl -v detach agent $A1
  # Return to attached
  iofogctl -v attach agent $A1
  checkAgent $A1
  checkDetachedAgentNegative $A1
  # Rename detached and succeed
  iofogctl -v rename agent $A1 $A0
  iofogctl -v rename agent $A0 albert --detached
  iofogctl -v detach agent $A0
  # Return to attached
  iofogctl -v attach agent $A0
  iofogctl -v rename agent $A0 $A1
  iofogctl -v rename agent albert $A0 --detached
}

@test "Attach agent" {
  local AGENT_NAME="${NAME}-0"
  iofogctl -v attach agent "$AGENT_NAME"
  checkAgent "$AGENT_NAME"
  checkDetachedAgentNegative "$AGENT_NAME"
}