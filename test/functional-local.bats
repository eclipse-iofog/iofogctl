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

@test "Deploy Agents against local Controller" {
  initLocalAgentFile
  test iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}_0"
}

@test "Deploy application" {
  initApplicationFiles
  test iofogctl -v -n "$NS" deploy application -f test/conf/application.yaml
  checkApplication
}

@test "Deploy microservice" {
  initMicroserviceFile
  test iofogctl -v -n "$NS" deploy microservice -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Update microservice" {
  initMicroserviceUpdateFile
  test iofogctl -v -n "$NS" deploy microservice -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
}

@test "Delete microservice" {
  test iofogctl -v -n "$NS" delete microservice "$MICROSERVICE_NAME"
  checkMicroserviceNegative
}

@test "Deploy microservice in application" {
  initMicroserviceFile
  test iofogctl -v -n "$NS" deploy microservice -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Deploy application from root file and test application update" {
  test iofogctl -v -n "$NS" deploy -f test/conf/root_application.yaml
  checkApplication
}

@test "Delete application" {
  test iofogctl -v -n "$NS" delete application "$APPLICATION_NAME"
  checkApplicationNegative
}

@test "Delete all" {
  test iofogctl -v -n "$NS" delete all
  checkControllerNegative
  checkConnectorNegative
  checkAgentNegative "${NAME}_0"
}

@test "Delete namespace" {
  test iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}