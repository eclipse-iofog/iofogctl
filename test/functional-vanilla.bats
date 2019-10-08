#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"

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
kind: iofog-controlplane
spec:
  controlplane:
    controllers:
    - name: $NAME
      user: $VANILLA_USER
      host: $VANILLA_HOST
      port: $VANILLA_PORT
      keyfile: $KEY_FILE
      version: $VANILLA_VERSION
      packagecloudtoken: $PACKAGE_CLOUD_TOKEN
    iofoguser:
      name: Testing
      surname: Functional
      email: user@domain.com
      password: S5gYVgLEZV
---
kind: iofog-connector
spec:
  connectors:
  - name: $NAME
    user: $VANILLA_USER
    host: $VANILLA_HOST
    port: $VANILLA_PORT
    keyfile: $KEY_FILE
    version: $VANILLA_VERSION
    packagecloudtoken: $PACKAGE_CLOUD_TOKEN" > test/conf/vanilla.yaml

  test iofogctl -v -n "$NS" deploy -f test/conf/vanilla.yaml
  checkController
  checkConnector
}

@test "Controller legacy commands after vanilla deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
}

@test "Get Controller logs after vanilla deploy" {
  test iofogctl -v -n "$NS" logs controller "$NAME"
}

@test "Deploy Agents against vanilla Controller" {
  initAgentsFile
  test iofogctl -v -n "$NS" deploy -f test/conf/agents.yaml
  checkAgents
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy application from root file and test deploy idempotence" {
  test iofogctl -v -n "$NS" deploy -f test/conf/root_application.yaml
  checkApplication
}

# Delete all does not delete application
@test "Delete application (bis)" {
  test iofogctl -v -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentsNegative
}

@test "Delete namespace" {
  test iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}