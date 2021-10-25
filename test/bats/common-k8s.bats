@test "Edge Resources" {
  startTest
  testEdgeResources
  stopTest
}

@test "Deploy Volumes" {
  startTest
  testDeployVolume
  testGetDescribeVolume
  stopTest
}

@test "Deploy Volumes Idempotent" {
  startTest
  testDeployVolume
  testGetDescribeVolume
  stopTest
}

@test "Delete Volumes and Redeploy" {
  startTest
  testDeleteVolume
  testDeployVolume
  testGetDescribeVolume
  stopTest
}

@test "Agent legacy commands" {
  startTest
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
  stopTest
}

@test "Get Agent logs" {
  startTest
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
  done
  stopTest
}

@test "Prune Agent" {
  startTest
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
  stopTest
}

@test "Disconnect from cluster" {
  startTest
  initAgents
  iofogctl -v -n "$NS" disconnect
  checkNamespaceExistsNegative "$NS"
  stopTest
}

@test "Connect to cluster using deploy file" {
  startTest
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect -f test/conf/k8s.yaml
  checkControllerK8s
  checkAgents
  stopTest
}

@test "Generate connection string" {
  startTest
  local IP=$(kctl get svc -l name=controller -n "$NS" | awk 'FNR > 1 {print $4}')
  testGenerateConnectionString "http://$IP:51121" # Disable this on local run
  CNCT=$(iofogctl -n "$NS" connect --generate)
  eval "$CNCT -n ${NS}-2"
  iofogctl disconnect -n "${NS}-2"
  stopTest
}

@test "Disconnect from cluster again" {
  startTest
  initAgents
  iofogctl -v -n "$NS" disconnect
  checkNamespaceExistsNegative "$NS"
  iofogctl -v -n "$NS2" disconnect # Idempotent
  stopTest
}

@test "Connect to cluster using flags" {
  startTest
  CONTROLLER_ENDPOINT=$(cat /tmp/endpoint.txt)
  iofogctl -v -n "$NS" connect --kube "$KUBE_CONFIG" --email "$USER_EMAIL" --pass "$USER_PW"
  checkControllerK8s
  checkAgents
  stopTest
}


@test "Set default namespace" {
  startTest
  testDefaultNamespace "$NS"
  stopTest
}

@test "Deploy application template" {
  startTest
  testApplicationTemplates
  stopTest
}

@test "Deploy application" {
  startTest
  initApplicationFiles
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Deploy route" {
  startTest
  initRouteFile
  iofogctl -v -n "$NS" deploy -f test/conf/route.yaml
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Volumes are mounted" {
  startTest
  testMountVolume
  stopTest
}

@test "Test Public Ports w/ Microservices on same Agent" {
  startTest
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
  PUBLIC_ENDPOINT=$(iofogctl -n "$NS" -v describe microservice $APPLICATION_NAME/"$MSVC2_NAME" | grep "\- http://" | sed 's|.*http://|http://|g')
  echo "PUBLIC_ENDPOINT: $PUBLIC_ENDPOINT"
  hitMsvcEndpoint "$PUBLIC_ENDPOINT"
  stopTest
}

@test "Move microservice to another agent" {
  startTest
  iofogctl -v move microservice $APPLICATION_NAME/$MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
  initAgents
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  waitForProxyMsvc ${HOSTS[1]} ${USERS[1]} $SSH_KEY_PATH
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Test Public Ports w/ Microservice on different Agents" {
  startTest
  # Wait for k8s service
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  testDefaultProxyConfig "$EXT_IP"
  PUBLIC_ENDPOINT=$(iofogctl -n "$NS" -v describe microservice $APPLICATION_NAME/"$MSVC2_NAME" | grep "\- http://" | sed 's|.*http://|http://|g')
  echo "PUBLIC_ENDPOINT: $PUBLIC_ENDPOINT"
  hitMsvcEndpoint "$PUBLIC_ENDPOINT"
  stopTest
}

@test "Change Microservice Ports" {
  startTest
  initApplicationFiles
  initApplicationWithPortFile
  EXT_IP=$(waitForSvc "$NS" http-proxy)
  iofogctl -v deploy -f test/conf/application.yaml
  # Wait for port to update to 6666
  PORT=""
  SECS=0
  MAX=60
  while [[ $SECS -lt $MAX && ! -z $PORT ]]; do
    PORT=$(kctl describe svc http-proxy -n "$NS" | grep 6666/TCP)
    SECS=$((SECS+1))
    sleep 1
  done
  # Check what happened in the loop above
  [ $SECS -lt $MAX ]
  waitForSvc "$NS" http-proxy
  stopTest
}

@test "Delete Public Port" {
  startTest
  initApplicationFiles
  initApplicationWithoutPortsFiles

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
  stopTest
}

# Delete all does not delete application
@test "Delete application" {
  startTest
  iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
  stopTest
}

@test "Deploy Agents for idempotence" {
  startTest
  initRemoteAgentsFile
  iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
  stopTest
}

@test "Configure Agents" {
  startTest
  initAgents
  iofogctl -v -n "$NS" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME"
  done
  stopTest
}

@test "Detach agent" {
  startTest
  local AGENT_NAME="${NAME}-0"
  iofogctl -v detach agent "$AGENT_NAME"
  checkAgentNegative "$AGENT_NAME"
  checkDetachedAgent "$AGENT_NAME"
  stopTest
}

@test "Detach with same name" {
  startTest
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
  stopTest
}

@test "Attach agent" {
  startTest
  local AGENT_NAME="${NAME}-0"
  iofogctl -v attach agent "$AGENT_NAME"
  checkAgent "$AGENT_NAME"
  checkDetachedAgentNegative "$AGENT_NAME"
  stopTest
}

@test "Move Agent" {
  startTest
  local AGENT_NAME="${NAME}-0"
  iofogctl -v -n "$NS" move "$AGENT_NAME" "$NS"
  iofogctl -v -n "$NS" get agents | grep "$AGENT_NAME"
  stopTest
}
