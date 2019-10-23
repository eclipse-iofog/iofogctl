#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"
NS2="$NS"_2

# TODO: Enable this when a release of Controller is usable here (version needs to be specified for dev package)
#@test "Deploy vanilla Controller" {
#  initVanillaController
#  test iofogctl -v -n "$NS" deploy controller "$NAME" --user "$VANILLA_USER" --host "$VANILLA_HOST" --key-file "$KEY_FILE" --port "$VANILLA_PORT"
#  checkController
#}

@test "Create namespace" {
  test iofogctl create namespace "$NS"
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
    user: $VANILLA_USER
    host: $VANILLA_HOST
    port: $VANILLA_PORT
    keyFile: $KEY_FILE
    version: $VANILLA_VERSION
    packageCloudToken: $PACKAGE_CLOUD_TOKEN
  iofogUser:
    name: Testing
    surname: Functional
    email: user@domain.com
    password: S5gYVgLEZV
---
apiVersion: iofog.org/v1
kind: Connector
metadata:
  name: $NAME
spec:
  user: $VANILLA_USER
  host: $VANILLA_HOST
  port: $VANILLA_PORT
  keyFile: $KEY_FILE
  version: $VANILLA_VERSION
  packageCloudToken: $PACKAGE_CLOUD_TOKEN" > test/conf/vanilla.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/vanilla.yaml
  checkController
  checkConnector
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Connector legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy connector "$NAME" status
  checkLegacyConnector
}

@test "Get Controller logs after vanilla deploy" {
  test iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
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

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy application and test deploy idempotence" {
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Connect in another namespace" {
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

# Delete all does not delete application
@test "Delete application" {
  test iofogctl -v -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Delete namespaces" {
  test iofogctl delete namespace "$NS"
  test iofogctl delete namespace "$NS2"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}