#!/usr/bin/env bash

. test/func/include.bash

NS="$NAMESPACE"

@test "Create namespace" {
  iofogctl create namespace "$NS"
}

@test "Test no executors" {
  testNoExecutors
}

@test "Test wrong namespace metadata " {
  testWrongNamespace
}

@test "Deploy local Controller" {
  initLocalControllerFile
  iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkControllerLocal
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
  checkControllerLocal
}

@test "Deploy Agents against local Controller again for indempotence" {
  initLocalAgentFile
  iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
}

@test "Deploy Volumes" {
  testDeployLocalVolume
  testGetDescribeLocalVolume
}

@test "Deploy Volumes Idempotent" {
  testDeployLocalVolume
  testGetDescribeLocalVolume
}

@test "Delete Volumes and Redeploy" {
  testDeleteLocalVolume
  testDeployLocalVolume
  testGetDescribeLocalVolume
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

@test "Deploy route" {
  initRouteFile
  iofogctl -v -n "$NS" deploy -f test/conf/route.yaml
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
}

@test "Update microservice" {
  initMicroserviceUpdateFile
  iofogctl -v -n "$NS" deploy -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
}

@test "Rename and Delete route" {
  local NEW_ROUTE_NAME="route-2"
  iofogctl -v -n "$NS" rename route "$ROUTE_NAME" "$NEW_ROUTE_NAME"
  iofogctl -v -n "$NS" delete route "$NEW_ROUTE_NAME"
  checkRouteNegative "$NEW_ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  checkRouteNegative "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
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

@test "Detach should fail because of running msvc" {
  run iofogctl -v -n "$NS" detach agent ${NAME}-0
  [ "$status" -eq 1 ]
  echo "$output" | grep "because it still has microservices running. Remove the microservices first, or use the --force option."
}

@test "Detach/attach agent" {
  iofogctl -v -n "$NS" detach agent ${NAME}-0 --force
  iofogctl -v -n "$NS" attach agent ${NAME}-0
}

@test "Deploy application from file and test application update" {
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
}

@test "Delete agent should fail because of running msvc" {
  run iofogctl -v -n "$NS" delete agent ${NAME}-0
  [ "$status" -eq 1 ]
  echo "$output" | grep "because it still has microservices running. Remove the microservices first, or use the --force option."
}

@test "Delete agent should work with --force option" {
  iofogctl -v -n "$NS" delete agent ${NAME}-0 --force
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
