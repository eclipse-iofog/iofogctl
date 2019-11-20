#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"

@test "Create namespace" {
  test iofogctl create namespace "$NS"
}

@test "Deploy local Controller" {
  initLocalControllerFile
  test iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkController
  checkConnector
}

@test "Controller legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Connector legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy connector "$NAME" status
  checkLegacyConnector
}

@test "Deploy Agents against local Controller" {
  initLocalAgentFile
  test iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
}

@test "Agent legacy commands" {
  test iofogctl -v -n "$NS" legacy agent "${NAME}-0" status
  checkLegacyAgent "${NAME}-0"
}

@test "Deploy local Controller again for indempotence" {
  initLocalControllerFile
  test iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkController
  checkConnector
}

@test "Deploy Agents against local Controller again for indempotence" {
  initLocalAgentFile
  test iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Deploy microservice" {
  initMicroserviceFile
  test iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Update microservice" {
  initMicroserviceUpdateFile
  test iofogctl -v -n "$NS" deploy -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
}

@test "Delete microservice using file option" {
  test iofogctl -v -n "$NS" delete -f test/conf/updatedMicroservice.yaml
  checkMicroserviceNegative
}

@test "Deploy microservice in application" {
  initMicroserviceFile
  test iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Deploy application from file and test application update" {
  test iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Delete all using file" {
  initAllLocalDeleteFile
  test iofogctl -v -n "$NS" delete -f test/conf/all-local.yaml
  checkApplicationNegative
  checkControllerNegative
  checkConnectorNegative
  checkAgentNegative "${NAME}-0"
}

@test "Delete namespace" {
  test iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}