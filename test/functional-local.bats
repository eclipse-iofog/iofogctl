#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS=$(echo "$NAMESPACE""-local")

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy local Controller" {
  echo "controlplane:
  controllers:
  - name: $NAME
    host: 127.0.0.1
    version: $VANILLA_VERSION
    packagecloudtoken: $PACKAGE_CLOUD_TOKEN
    iofoguser:
      name: Testing
      surname: Functional
      email: user@domain.com
      password: S5gYVgLEZV" > test/conf/local.yaml

  test iofogctl -q -n "$NS" deploy -f test/conf/local.yaml
  checkController
}

@test "Deploy Agents against local Controller" {
  initLocalAgentFile
  test iofogctl -q -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}_0"
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -q -n "$NS" deploy application -f test/conf/application.yaml
  checkApplication
}

@test "Delete application" {
  test iofogctl -q -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Deploy application from root file" {
  test iofogctl -q -n "$NS" deploy -f test/conf/root_application.yaml
  checkApplication
}

# Delete all does not delete application
@test "Delete application (bis)" {
  test iofogctl -q -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -q -n "$NS" delete all
  checkControllerNegative
  checkAgentNegative "${NAME}_0"
}

@test "Delete namespace" {
  test iofogctl delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}