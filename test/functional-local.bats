#!/usr/bin/env bash

. test/functions.bash
. test/functional.vars.bash

NS="$NAMESPACE"

@test "Create namespace" {
  iofogctl create namespace "$NS"
}

@test "Deploy local Controller" {
  initLocalControllerFile
  iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkController
}

@test "Controller legacy commands after deploy" {
  iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
}

@test "Deploy Agents against local Controller" {
  initLocalAgentFile
  iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
}

@test "Agent legacy commands" {
  iofogctl -v -n "$NS" legacy agent "${NAME}-0" status
  checkLegacyAgent "${NAME}-0"
}

@test "Deploy local Controller again for indempotence" {
  initLocalControllerFile
  iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkController
}

@test "Deploy Agents against local Controller again for indempotence" {
  initLocalAgentFile
  iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
}

@test "Deploy application" {
  initApplicationFiles
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
}

@test "Deploy microservice" {
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Update microservice" {
  initMicroserviceUpdateFile
  iofogctl -v -n "$NS" deploy -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
}

@test "Delete microservice using file option" {
  iofogctl -v -n "$NS" delete -f test/conf/updatedMicroservice.yaml
  checkMicroserviceNegative
}

@test "Deploy microservice in application" {
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
}

@test "Deploy application from file and test application update" {
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Get local logs" {
  [[ ! -z $(iofogctl -v -n "$NS" logs controller $NAME) ]]
  [[ ! -z $(iofogctl -v -n "$NS" logs agent ${NAME}-0) ]]
}

@test "Deploy registry" {
  initGCRRegistryFile
  iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkGCRRegistry
  initUpdatedGCRRegistryFile
  iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkUpdatedGCRRegistry
  iofogctl -v -n "$NS" delete registry 3
  checkGCRRegistryNegative
}

@test "Delete all using file" {
  initAllLocalDeleteFile
  iofogctl -v -n "$NS" delete -f test/conf/all-local.yaml
  checkApplicationNegative
  checkControllerNegative
  checkAgentNegative "${NAME}-0"
}

@test "Delete namespace" {
  iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
}
