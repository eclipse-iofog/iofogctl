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
}

@test "Controller legacy commands after deploy" {
  test iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
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
  waitForMsvc func-app-server "$NS"
  waitForMsvc func-app-ui "$NS"
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

@test "Get local logs" {
  [[ ! -z $(iofogctl -v -n "$NS" logs controller $NAME) ]]
  [[ ! -z $(iofogctl -v -n "$NS" logs agent ${NAME}-0) ]]
}

@test "Deploy registry" {
  initGCRRegistryFile
  test iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkGCRRegistry
  initUpdatedGCRRegistryFile
  test iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkUpdatedGCRRegistry
  test iofogctl -v -n "$NS" delete registry 3
  checkGCRRegistryNegative
}

@test "Delete all using file" {
  initAllLocalDeleteFile
  test iofogctl -v -n "$NS" delete -f test/conf/all-local.yaml
  checkLocalResourcesDeleted
  checkApplicationNegative
  checkControllerNegative
  checkAgentNegative "${NAME}-0"
}

@test "Delete namespace" {
  test iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
