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
      version: $VANILLA_VERSION
  iofogUser:
    name: Testing
    surname: Functional
    email: $USER_EMAIL
    password: $USER_PW
---
apiVersion: iofog.org/v1
kind: Connector
metadata:
  name: $NAME
spec:
  host: $VANILLA_HOST
  ssh:
    user: $VANILLA_USER
    port: $VANILLA_PORT
    keyFile: $KEY_FILE
  package:
    version: $VANILLA_VERSION" > test/conf/vanilla.yaml

  test iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  checkConnector
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -v legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Connector legacy commands after deploy" {
  test iofogctl -v legacy connector "$NAME" status
  checkLegacyConnector
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

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy application and test deploy idempotence" {
  test iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
}

@test "Connect in another namespace using file" {
  test iofogctl -v -n "$NS2" connect -f test/conf/vanilla.yaml
  checkController "$NS2"
  checkConnector "$NS2"
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
  checkConnectorNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}

@test "Connect in other namespace using flags" {
  initVanillaController
  CONTROLLER_ENDPOINT="$VANILLA_HOST:51121"
  test iofogctl -v -n "$NS2" connect --name "$NAME" --endpoint "$CONTROLLER_ENDPOINT" --email "$USER_EMAIL" --pass "$USER_PW"
  checkController "$NS2"
  checkConnector "$NS2"
  checkAgents "$NS2"
}

@test "Configure Controller and Connector" {
  initVanillaController
  for resource in controller connector; do
    test iofogctl -v -n "$NS2" configure "$resource" "$NAME" --host "$VANILLA_HOST" --user "$VANILLA_USER" --port "$VANILLA_PORT" --key "$KEY_FILE"
  done
  test iofogctl -v -n "$NS2" logs controller "$NAME"
  checkLegacyController "$NS2"
  checkLegacyConnector "$NS2"
}

@test "Configure Agents" {
  initAgents
  echo "========> Config.yaml before configure"
  cat ~/.iofog/config.yaml
  test iofogctl -v -n "$NS2" configure agents --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
  for IDX in "${!AGENTS[@]}"; do
    local AGENT_NAME="${NAME}-${IDX}"
    test iofogctl -v -n "$NS2" configure agent "$AGENT_NAME" --port "${PORTS[IDX]}" --key "$KEY_FILE" --user "${USERS[IDX]}"
    echo "========> Config.yaml"
    cat ~/.iofog/config.yaml
    test iofogctl -v -n "$NS2" logs agent "$AGENT_NAME"
    checkLegacyAgent "$AGENT_NAME" "$NS2"
  done
}

@test "Rename Agents" {
  initAgents
  test iofogctl -v -n "$NS2" rename agent "${NAME}" "${NAME}-newname"
  checkRenamedResource agent "${NAME}" "${NAME}-newname"
  test iofogctl -v -n "$NS2" rename agent "${NAME}-newname" "${NAME}"
  checkRenamedResource agent "${NAME}-newname" "${NAME}"
}

@test "Rename Controller" {
  initVanillaController
  test iofogctl -v -n "$NS2" rename controller "${NAME}" "${NAME}-newname"
  checkRenamedResource controller "${NAME}" "${NAME}-newname"
  test iofogctl -v -n "$NS2" rename controller "${NAME}-newname" "${NAME}"
  checkRenamedResource controller "${NAME}-newname" "${NAME}"
}

@test "Rename Connector" {
  initVanillaController
  test iofogctl -v -n "$NS2" rename connector "${NAME}" "${NAME}-newname"
  checkRenamedResource connector "${NAME}" "${NAME}-newname"
  test iofogctl -v -n "$NS2" rename connector "${NAME}-newname" "${NAME}"
  checkRenamedResource connector "${NAME}-newname" "${NAME}"
}


@test "Rename Namespace" {
  test iofogctl -v -n "$NS2" rename namespace "${NS2}" "${NS2}-newname"
  checkRenamedResource namespace "${NS2}" "${NS2}-newname"
  test iofogctl -v -n "$NS2" rename namespace "${NS2}-newname" "${NS2}"
  checkRenamedResource namespace "${NS2}-newname" "${NS2}"
}


@test "Disconnect other namespace again" {
  test iofogctl -v -n "$NS2" disconnect
  checkControllerNegative "$NS2"
  checkConnectorNegative "$NS2"
  checkAgentsNegative "$NS2"
  checkApplicationNegative "$NS2"
}


@test "Deploy again to check it doesn't lose database" {
  test iofogctl -v deploy -f test/conf/vanilla.yaml
  checkController
  checkConnector
  initAgentsFile
  test iofogctl -v deploy -f test/conf/agents.yaml
  checkAgents
  checkApplication
}

@test "Rename Application" {
  test iofogctl -v rename application "${APPLICATION_NAME}" "${APPLICATION_NAME}-newname"
  checkRenamedResource application "${APPLICATION_NAME}" "${APPLICATION_NAME}-newname"
  test iofogctl -v rename application "${APPLICATION_NAME}-newname" "${APPLICATION_NAME}"
  checkRenamedResource application "${APPLICATION_NAME}-newname" "${APPLICATION_NAME}"
}

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -v delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Delete namespaces" {
  test iofogctl delete namespace "$NS"
  test iofogctl delete namespace "$NS2"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}