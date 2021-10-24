#!/usr/bin/env bash

. test/func/include.bash

NS="$NAMESPACE"
NS2="$NS"-2

@test "Initialize tests" {
  stopTest
}

@test "Verify Agents >= 2" {
  startTest
  testAgentCount
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

@test "Set default namespace" {
  startTest
  testDefaultNamespace "$NS"
  stopTest
}

@test "Deploy vanilla Controller" {
  startTest
  initVanillaController
  echo "---
apiVersion: iofog.org/v3
kind: ControlPlane
metadata:
  name: func-controlplane
spec:
  controllers:
  - name: $NAME
    host: $VANILLA_HOST
    ssh:
      user: $VANILLA_USER
      port: $VANILLA_PORT
      keyFile: $KEY_FILE
  package:
    repo: $CONTROLLER_REPO
    version: $CONTROLLER_VANILLA_VERSION
    token: $CONTROLLER_PACKAGE_CLOUD_TOKEN
  systemAgent:
    repo: $AGENT_REPO
    version: $AGENT_VANILLA_VERSION
    token: $AGENT_PACKAGE_CLOUD_TOKEN
  systemMicroservices:
    router:
      x86: $ROUTER_IMAGE
      arm: $ROUTER_ARM_IMAGE
    proxy:
      x86: $PROXY_IMAGE
      arm: $PROXY_ARM_IMAGE
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW" > test/conf/vanilla.yaml

  iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  stopTest
}

@test "Check Controller host has a system Agent running on it with qrouter microservice" {
  startTest
  initVanillaController

  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $SSH_KEY_PATH $VANILLA_USER@$VANILLA_HOST"
  [[ "ok" == $($SSH_COMMAND -- sudo iofog-agent status | grep 'Controller' | awk '{print $5}') ]]
  [[ "RUNNING" == $($SSH_COMMAND --  sudo iofog-agent status | grep 'daemon' | awk '{print $4}') ]]
  [[ "http://${VANILLA_HOST}:51121/api/v3/" == $($SSH_COMMAND -- sudo iofog-agent info | grep 'Controller' | awk '{print $4}') ]]
  [[ $($SSH_COMMAND -- sudo cat /etc/iofog-agent/microservices.json | grep "router") ]]
  stopTest
}

@test "Controller legacy commands after vanilla deploy" {
  startTest
  iofogctl -v legacy controller "$NAME" iofog list
  checkLegacyController
  stopTest
}

@test "Get Controller logs after vanilla deploy" {
  startTest
  iofogctl -v logs controller "$NAME"
  stopTest
}

@test "Deploy Agents against vanilla Controller" {
  startTest
  initRemoteAgentsFile
  iofogctl --debug deploy -f test/conf/agents.yaml
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
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
  stopTest
}

@test "Prune Agent" {
  startTest
  initVanillaController
  initAgents
  local AGENT_NAME="${NAME}-0"
  iofogctl -v prune agent "$AGENT_NAME"
  local CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  echo "$CONTROLLER_ENDPOINT"
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  # TODO: Enable check that is not flake
  #checkAgentPruneController "$CONTROLLER_ENDPOINT" "$SSH_KEY_PATH"
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

@test "Update detached agent name" {
  startTest
  local OLD_NAME="${NAME}-0"
  local NEW_NAME="${NAME}-renamed"
  iofogctl -v rename agent "$OLD_NAME" "$NEW_NAME" --detached
  checkDetachedAgentNegative "$OLD_NAME"
  checkDetachedAgent "$NEW_NAME"
  iofogctl -v rename agent "$NEW_NAME" "$OLD_NAME" --detached
  checkDetachedAgentNegative "$NEW_NAME"
  checkDetachedAgent "$OLD_NAME"
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

@test "Deploy application template" {
  startTest
  testApplicationTemplates
  stopTest
}

@test "Deploy application" {
  startTest
  initApplicationFileWithRoutes
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Volumes are mounted" {
  startTest
  testMountVolume
  stopTest
}

@test "Microservice logs" {
  startTest
  iofogctl -v logs microservice "$APPLICATION_NAME"/"$MSVC2_NAME" | grep "node index.js"
  stopTest
}

@test "Deploy application and test deploy idempotence" {
  startTest
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Test Public Ports w/ Microservices on same Agent" {
  startTest
  initVanillaController
  initAgents
  # Wait for proxy microservice
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  # Wait for proxy microservice
  waitForProxyMsvc ${HOSTS[0]} ${USERS[0]} $SSH_KEY_PATH
  waitForProxyMsvc $VANILLA_HOST $VANILLA_USER $SSH_KEY_PATH
  # Hit the endpoint
  EXT_IP=$VANILLA_HOST
  testDefaultProxyConfig "$EXT_IP"
  PUBLIC_ENDPOINT=$(iofogctl -n "$NS" -v describe microservice $APPLICATION_NAME/"$MSVC2_NAME" | grep "\- http://" | sed 's|.*http://|http://|g')
  echo "PUBLIC_ENDPOINT: $PUBLIC_ENDPOINT"
  hitMsvcEndpoint "$PUBLIC_ENDPOINT"
  stopTest
}

@test "Move microservice to another agent" {
  startTest
  iofogctl -v move microservice $APPLICATION_NAME/$MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
  # Avoid checking RUNNING state of msvc on first agent
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Test Public Ports w/ Microservice on different Agents" {
  startTest
  initVanillaController
  initAgents
  # Wait for proxy microservice
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  waitForProxyMsvc ${HOSTS[1]} ${USERS[1]} $SSH_KEY_PATH
  waitForProxyMsvc $VANILLA_HOST $VANILLA_USER $SSH_KEY_PATH
  # Hit the endpoint
  EXT_IP=$VANILLA_HOST
  testDefaultProxyConfig "$EXT_IP"
  PUBLIC_ENDPOINT=$(iofogctl -n "$NS" -v describe microservice $APPLICATION_NAME/"$MSVC2_NAME" | grep "\- http://" | sed 's|.*http://|http://|g')
  echo "PUBLIC_ENDPOINT: $PUBLIC_ENDPOINT"
  hitMsvcEndpoint "$PUBLIC_ENDPOINT"
  stopTest
}

@test "Generate connection string" {
  startTest
  initVanillaController
  testGenerateConnectionString "http://$VANILLA_HOST:51121"
  CNCT=$(iofogctl -n "$NS" connect --generate)
  eval "$CNCT -n ${NS}-2"
  iofogctl disconnect -n "${NS}-2"
  stopTest
}

@test "Connect in another namespace using file" {
  startTest
  iofogctl -v -n "$NS2" connect -f test/conf/vanilla.yaml
  checkControllerAfterConnect "$NS2"
  checkAgents "$NS2"
  checkApplication "$NS2"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" legacy agent "$AGENT_NAME" status
  done
  stopTest
}

@test "Test you can access logs in other namespace" {
  startTest
  initVanillaController
  iofogctl -v -n "$NS2" configure controller "$NAME" --user "$VANILLA_USER" --key "$KEY_FILE" --port $VANILLA_PORT
  checkControllerAfterConfigure "$NS2"
  iofogctl -v -n "$NS2" logs controller "$NAME"
  stopTest
}

@test "Disconnect other namespace" {
  startTest
  iofogctl -v -n "$NS2" disconnect
  checkNamespaceExistsNegative "$NS2"
  iofogctl -v -n "$NS2" disconnect # Idempotent
  stopTest
}

@test "Connect in other namespace using flags" {
  startTest
  initVanillaController
  CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  iofogctl -v -n "$NS2" connect --name "$NAME" --ecn-addr "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkControllerAfterConnect "$NS2"
  checkAgents "$NS2"
  stopTest
}

@test "Configure Controller" {
  startTest
  initVanillaController
  iofogctl -v -n "$NS2" configure controller "$NAME" --user "$VANILLA_USER" --port $VANILLA_PORT --key "$KEY_FILE"
  checkControllerAfterConfigure "$NS2"
  iofogctl -v -n "$NS2" logs controller "$NAME"

  iofogctl -v -n "$NS2" configure controllers "$NAME" --user "$VANILLA_USER" --port $VANILLA_PORT --key "$KEY_FILE"
  checkControllerAfterConfigure "$NS2"
  iofogctl -v -n "$NS2" logs controller "$NAME"
  stopTest
}

@test "Configure Agents" {
  startTest
  initAgents
  iofogctl -v -n "$NS2" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME" "$NS2"
  done
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" configure agent "$AGENT_NAME" --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
    iofogctl -v -n "$NS2" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME" "$NS2"
  done
  stopTest
}

@test "Rename Agents" {
  startTest
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" rename agent "$AGENT_NAME" "newname"
    checkRenamedResource agents "$AGENT_NAME" "newname" "$NS2"
    iofogctl -v -n "$NS2" rename agent "newname" "$AGENT_NAME"
    checkRenamedResource agents "newname" "$AGENT_NAME" "$NS2"
  done
  stopTest
}

@test "Rename Controller" {
  startTest
  iofogctl -v -n "$NS2" rename controller "$NAME" "newname"
  checkRenamedResource controllers "$NAME" "newname" "$NS2"
  iofogctl -v -n "$NS2" rename controller "newname" "$NAME"
  checkRenamedResource controllers "newname" "$NAME" "$NS2"
  stopTest
}

@test "Rename Namespace" {
  startTest
  iofogctl -v rename namespace "${NS2}" "newname"
  checkRenamedNamespace "$NS2" "newname"
  iofogctl -v rename namespace "newname" "${NS2}"
  checkRenamedNamespace "newname" "$NS2"
  stopTest
}

@test "Rename Application" {
  startTest
  iofogctl -v rename application "$APPLICATION_NAME" "application-name"
  iofogctl get all
  checkRenamedApplication "$APPLICATION_NAME" "application-name" "$NS"
  iofogctl -v rename application "application-name" "$APPLICATION_NAME"
  checkRenamedApplication "application-name" "$APPLICATION_NAME" "$NS"
  stopTest
}

@test "Disconnect other namespace again" {
  startTest
  iofogctl -v -n "$NS2" disconnect
  checkNamespaceExistsNegative "$NS2"
  stopTest
}

@test "Deploy again to check it doesn't lose database" {
  startTest
  iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  initRemoteAgentsFile
  iofogctl -v deploy -f test/conf/agents.yaml
  checkAgents
  checkApplication
  stopTest
}

# Delete all does not delete application
@test "Delete application" {
  startTest
  iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
  stopTest
}

@test "Delete all" {
  startTest
  iofogctl -v delete all
  initVanillaController
  checkVanillaResourceDeleted $VANILLA_USER $VANILLA_HOST $VANILLA_PORT $KEY_FILE "iofog-controller"
 
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    checkVanillaResourceDeleted ${USERS[$IDX]} ${HOSTS[$IDX]} ${PORTS[$IDX]} $KEY_FILE "iofog-agent"
  done

  checkControllerNegative
  checkAgentsNegative
  stopTest
}

@test "Delete namespaces" {
  startTest
  iofogctl delete namespace "$NS"
  checkNamespaceExistsNegative "$NS"
  stopTest
}
