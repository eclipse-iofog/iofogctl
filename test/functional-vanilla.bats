#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"
NS2="$NS"_2
USER_PW="S5gYVgLEZV"
USER_EMAIL="user@domain.com"

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Set default namespace" {
  test iofogctl configure default-namespace "$NS"
}

@test "Deploy vanilla Controller" {
  initVanillaController
  echo "---
apiVersion: iofog.org/v1
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
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW" > test/conf/vanilla.yaml

  test iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
}

@test "Check Controller host has a system Agent running on it with qrouter microservice" {
  initVanillaController
  SSH_COMMAND="ssh -oStrictHostKeyChecking=no -i $KEY_FILE $VANILLA_USER@$VANILLA_HOST"
  [[ "ok" == $($SSH_COMMAND -- sudo iofog-agent status | grep 'Controller' | awk '{print $5}') ]]
  [[ "RUNNING" == $($SSH_COMMAND --  sudo iofog-agent status | grep 'daemon' | awk '{print $4}') ]]
  [[ "http://${VANILLA_HOST}:51121/api/v3/" == $($SSH_COMMAND -- sudo iofog-agent info | grep 'Controller' | awk '{print $4}') ]]
  $SSH_COMMAND -- sudo cat /etc/iofog-agent/microservices.json
  $SSH_COMMAND -- sudo cat /etc/iofog-agent/microservices.json | jq '.data[0].imageId'
  [[ "\"quay.io/interconnectedcloud/qdrouterd:latest\"" == $($SSH_COMMAND -- sudo cat /etc/iofog-agent/microservices.json | jq '.data[0].imageId') ]]
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -v legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Get Controller logs after vanilla deploy" {
  test iofogctl -v logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
  initAgentsFile
  test iofogctl -v deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Agent legacy commands" {
  initAgents
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v legacy agent "$AGENT_NAME" status
    checkLegacyAgent "$AGENT_NAME"
  done
}

@test "Prune Agent" {
  initVanillaController
  initAgents
  local AGENT_NAME="${NAME}-0"
  test iofogctl -v prune agent "$AGENT_NAME"
  local CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  echo "$CONTROLLER_ENDPOINT"
  local SSH_KEY_PATH=$KEY_FILE
  if [[ ! -z $WSL_KEY_FILE ]]; then
    SSH_KEY_PATH=$WSL_KEY_FILE
  fi
  # TODO: Enable check that is not flake
  #checkAgentPruneController "$CONTROLLER_ENDPOINT" "$SSH_KEY_PATH"
}

# @test "Detach agent" {
#   local AGENT_NAME="${NAME}-0"
#   test iofogctl -v detach agent "$AGENT_NAME"
#   checkAgentNegative "$AGENT_NAME"
#   checkDetachedAgent "$AGENT_NAME"
# }

# @test "Update detached agent name" {
#   local OLD_NAME="${NAME}-0"
#   local NEW_NAME="${NAME}-renamed"
#   test iofogctl -v rename agent "$OLD_NAME" "$NEW_NAME" --detached
#   checkDetachedAgentNegative "$OLD_NAME"
#   checkDetachedAgent "$NEW_NAME"
#   test iofogctl -v rename agent "$NEW_NAME" "$OLD_NAME" --detached
#   checkDetachedAgentNegative "$NEW_NAME"
#   checkDetachedAgent "$OLD_NAME"
# }

# @test "Attach agent" {
#   local AGENT_NAME="${NAME}-0"
#   test iofogctl -v attach agent "$AGENT_NAME"
#   checkAgent "$AGENT_NAME"
#   checkDetachedAgentNegative "$AGENT_NAME"
# }

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy application and test deploy idempotence" {
  test iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc func-app-server "$NS"
  waitForMsvc func-app-ui "$NS"
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
  EXT_IP=$VANILLA_HOST
  # Hit the endpoint
  COUNT=$(curl --max-time 120 http://${EXT_IP}:5000/api/raw | jq '. | length')
  [ $COUNT -gt 0 ]
}

@test "Move microservice to another agent" {
  test iofogctl -v move microservice $MSVC2_NAME ${NAME}-1
  checkMovedMicroservice $MSVC2_NAME ${NAME}-1
  # Avoid checking RUNNING state of msvc on first agent
  waitForMsvc "$MSVC2_NAME" "$NS" "DELETING"
  waitForMsvc "$MSVC2_NAME" "$NS"
}

# TODO: Enable when fixed
#@test "Test Public Ports w/ Microservice on different Agents" {
#  initVanillaController
#  initAgents
#  # Wait for proxy microservice
#  local SSH_KEY_PATH=$KEY_FILE
#  if [[ ! -z $WSL_KEY_FILE ]]; then
#    SSH_KEY_PATH=$WSL_KEY_FILE
#  fi
#  waitForProxyMsvc ${HOSTS[1]} ${USERS[1]} $SSH_KEY_PATH
#  EXT_IP=$VANILLA_HOST
#  # Hit the endpoint
#  COUNT=$(curl --max-time 120 http://${EXT_IP}:5000/api/raw | jq '. | length')
#  [ $COUNT -gt 0 ]
#}

@test "Connect in another namespace using file" {
  test iofogctl -v -n "$NS2" connect -f test/conf/vanilla.yaml
  checkController "$NS2"
  checkAgents "$NS2"
  checkApplication "$NS2"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS2" legacy agent "$AGENT_NAME" status
  done
}

@test "Disconnect other namespace" {
  test iofogctl -v -n "$NS2" disconnect
  checkControllerNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}

@test "Connect in other namespace using flags" {
  initVanillaController
  CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  test iofogctl -v -n "$NS2" connect --name "$NAME" --ecn-addr "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController "$NS2"
  checkAgents "$NS2"
}

@test "Configure Controller and Connector" {
  initVanillaController
  for resource in controller; do
    test iofogctl -v -n "$NS2" configure "$resource" "$NAME" --host "$VANILLA_HOST" --user "$VANILLA_USER" --port "$VANILLA_PORT" --key "$KEY_FILE"
  done
  test iofogctl -v -n "$NS2" logs controller "$NAME"
}

@test "Configure Agents" {
  initAgents
  test iofogctl -v -n "$NS2" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS2" configure agent "$AGENT_NAME" --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
    test iofogctl -v -n "$NS2" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME" "$NS2"
  done
}

@test "Rename Agents" {
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS2" rename agent "$AGENT_NAME" "newname"
    checkRenamedResource agents "$AGENT_NAME" "newname" "$NS2"
    test iofogctl -v -n "${NS2}" rename agent "newname" "${AGENT_NAME}"
    checkRenamedResource agents "newname" "$AGENT_NAME" "$NS2"
  done
}

@test "Rename Controller" {
  test iofogctl -v -n "$NS2" rename controller "$NAME" "newname"
  checkRenamedResource controllers "$NAME" "newname" "$NS2"
  test iofogctl -v -n "$NS2" rename controller "newname" "${NAME}"
  checkRenamedResource controllers "newname" "$NAME" "$NS2"
}

@test "Rename Namespace" {
  test iofogctl -v rename namespace "${NS2}" "newname"
  checkRenamedNamespace "$NS2" "newname"
  test iofogctl -v rename namespace "newname" "${NS2}"
  checkRenamedNamespace "newname" "$NS2"
}

@test "Rename Application" {
  test iofogctl -v rename application "$APPLICATION_NAME" "application-name"
  iofogctl get all
  checkRenamedApplication "$APPLICATION_NAME" "application-name" "$NS"
  test iofogctl -v rename application "application-name" "$APPLICATION_NAME"
  checkRenamedApplication "application-name" "$APPLICATION_NAME" "$NS"
}


@test "Disconnect other namespace again" {
  test iofogctl -v -n "$NS2" disconnect
  checkControllerNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}


@test "Deploy again to check it doesn't lose database" {
  test iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  initAgentsFile
  test iofogctl -v deploy -f test/conf/agents.yaml
  checkAgents
  checkApplication
}

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v delete all
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
  test iofogctl delete namespace "$NS"
  test iofogctl delete namespace "$NS2"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
