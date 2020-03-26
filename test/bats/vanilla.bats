#!/usr/bin/env bash

. test/func/include.bash

NS="$NAMESPACE"
NS2="$NS"_2

@test "Create namespace" {
  iofogctl create namespace "$NS"
}

@test "Set default namespace" {
  iofogctl configure default-namespace "$NS"
}

@test "Deploy vanilla Controller" {
  initVanillaController
  echo "---
apiVersion: iofog.org/v2
kind: RemoteControlPlane
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
      version: $AGENT_VANILLA_VERSION
      repo: $AGENT_REPO
      token: $AGENT_PACKAGE_CLOUD_TOKEN
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW" > test/conf/vanilla.yaml

  iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
}

@test "Check Controller host has a system Agent running on it with qrouter microservice" {
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
}

@test "Controller legacy commands after vanilla deploy" {
  iofogctl -v legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Get Controller logs after vanilla deploy" {
  iofogctl -v logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
  initRemoteAgentsFile
  iofogctl -v deploy -f test/conf/agents.yaml
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
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Prune Agent" {
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
}

@test "Detach agent" {
  local AGENT_NAME="${NAME}-0"
  iofogctl -v detach agent "$AGENT_NAME"
  checkAgentNegative "$AGENT_NAME"
  checkDetachedAgent "$AGENT_NAME"
}

@test "Update detached agent name" {
  local OLD_NAME="${NAME}-0"
  local NEW_NAME="${NAME}-renamed"
  iofogctl -v rename agent "$OLD_NAME" "$NEW_NAME" --detached
  checkDetachedAgentNegative "$OLD_NAME"
  checkDetachedAgent "$NEW_NAME"
  iofogctl -v rename agent "$NEW_NAME" "$OLD_NAME" --detached
  checkDetachedAgentNegative "$NEW_NAME"
  checkDetachedAgent "$OLD_NAME"
}

@test "Attach agent" {
  local AGENT_NAME="${NAME}-0"
  iofogctl -v attach agent "$AGENT_NAME"
  checkAgent "$AGENT_NAME"
  checkDetachedAgentNegative "$AGENT_NAME"
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

@test "Deploy application and test deploy idempotence" {
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
}

@test "Test Public Ports w/ Microservices on same Agent" {
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
  hitMsvcEndpoint "$EXT_IP"
}

@test "Move microservice to another agent" {
  iofogctl -v move microservice $MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
  # Avoid checking RUNNING state of msvc on first agent
  waitForMsvc "$MSVC2_NAME" "$NS"
}

@test "Test Public Ports w/ Microservice on different Agents" {
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
  hitMsvcEndpoint "$EXT_IP"
}

@test "Connect in another namespace using file" {
  iofogctl -v -n "$NS2" connect -f test/conf/vanilla.yaml
  checkController "$NS2"
  checkAgents "$NS2"
  checkApplication "$NS2"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" legacy agent "$AGENT_NAME" status
  done
}

@test "Disconnect other namespace" {
  iofogctl -v -n "$NS2" disconnect
  checkNamespaceExistsNegative "$NS2"
}

@test "Connect in other namespace using flags" {
  initVanillaController
  CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  iofogctl -v -n "$NS2" connect --name "$NAME" --ecn-addr "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController "$NS2"
  checkAgents "$NS2"
}

@test "Configure Controller and Connector" {
  initVanillaController
  for resource in controlplane; do
    iofogctl -v -n "$NS2" configure "$resource" "$NAME" --host "$VANILLA_HOST" --user "$VANILLA_USER" --port "$VANILLA_PORT" --key "$KEY_FILE"
  done
  iofogctl -v -n "$NS2" logs controller "$NAME"
}

@test "Configure Agents" {
  initAgents
  iofogctl -v -n "$NS2" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" configure agent "$AGENT_NAME" --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
    iofogctl -v -n "$NS2" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME" "$NS2"
  done
}

@test "Rename Agents" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    iofogctl -v -n "$NS2" rename agent "$AGENT_NAME" "newname"
    checkRenamedResource agents "$AGENT_NAME" "newname" "$NS2"
    iofogctl -v -n "${NS2}" rename agent "newname" "${AGENT_NAME}"
    checkRenamedResource agents "newname" "$AGENT_NAME" "$NS2"
  done
}

@test "Rename Controller" {
  iofogctl -v -n "$NS2" rename controller "$NAME" "newname"
  checkRenamedResource controllers "$NAME" "newname" "$NS2"
  iofogctl -v -n "$NS2" rename controller "newname" "${NAME}"
  checkRenamedResource controllers "newname" "$NAME" "$NS2"
}

@test "Rename Namespace" {
  iofogctl -v rename namespace "${NS2}" "newname"
  checkRenamedNamespace "$NS2" "newname"
  iofogctl -v rename namespace "newname" "${NS2}"
  checkRenamedNamespace "newname" "$NS2"
}

@test "Rename Application" {
  iofogctl -v rename application "$APPLICATION_NAME" "application-name"
  iofogctl get all
  checkRenamedApplication "$APPLICATION_NAME" "application-name" "$NS"
  iofogctl -v rename application "application-name" "$APPLICATION_NAME"
  checkRenamedApplication "application-name" "$APPLICATION_NAME" "$NS"
}


@test "Disconnect other namespace again" {
  iofogctl -v -n "$NS2" disconnect
  checkNamespaceExistsNegative "$NS2"
}


@test "Deploy again to check it doesn't lose database" {
  iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  initRemoteAgentsFile
  iofogctl -v deploy -f test/conf/agents.yaml
  checkAgents
  checkApplication
}

# Delete all does not delete application
@test "Delete application" {
  iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  iofogctl -v delete all
  initVanillaController
  checkVanillaResourceDeleted $VANILLA_USER $VANILLA_HOST $VANILLA_PORT $KEY_FILE "iofog-controller"
 
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    checkVanillaResourceDeleted ${USERS[$IDX]} ${HOSTS[$IDX]} ${PORTS[$IDX]} $KEY_FILE "iofog-agent"
  done

  checkControllerNegative
  checkAgentsNegative
}

@test "Delete namespaces" {
  iofogctl delete namespace "$NS"
  iofogctl delete namespace "$NS2"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
