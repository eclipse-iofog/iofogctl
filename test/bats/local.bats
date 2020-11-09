#!/usr/bin/env bash

. test/func/include.bash

NS="$NAMESPACE"

@test "Initialize tests" {
  stopTest
}

@test "Create namespace" {
  startTest
  iofogctl create namespace "$NS"
  stopTest
}

@test "Test no executors" {
  startTest
  testNoExecutors
  stopTest
}

@test "Test wrong namespace metadata " {
  startTest
  testWrongNamespace
  stopTest
}

@test "Deploy local Controller" {
  startTest
  initLocalControllerFile
  iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkControllerLocal
  stopTest
}

@test "Controller legacy commands after deploy" {
  startTest
  iofogctl -v -n "$NS" legacy controller "$NAME" iofog list
  checkLegacyController
  stopTest
}

@test "Deploy Agents against local Controller" {
  startTest
  initLocalAgentFile
  iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
  stopTest
}

@test "Edge Resources" {
  startTest
  testEdgeResources
  stopTest
}

@test "Agent legacy commands" {
  startTest
  iofogctl -v -n "$NS" legacy agent "${NAME}-0" status
  checkLegacyAgent "${NAME}-0"
  stopTest
}

@test "Deploy local Controller again for indempotence" {
  startTest
  initLocalControllerFile
  iofogctl -v -n "$NS" deploy -f test/conf/local.yaml
  checkControllerLocal
  stopTest
}

@test "Deploy Agents against local Controller again for indempotence" {
  startTest
  initLocalAgentFile
  iofogctl -v -n "$NS" deploy -f test/conf/local-agent.yaml
  checkAgent "${NAME}-0"
  stopTest
}

@test "Deploy Volumes" {
  startTest
  testDeployLocalVolume
  testGetDescribeLocalVolume
  stopTest
}

@test "Deploy Volumes Idempotent" {
  startTest
  testDeployLocalVolume
  testGetDescribeLocalVolume
  stopTest
}

@test "Delete Volumes and Redeploy" {
  startTest
  testDeleteLocalVolume
  testDeployLocalVolume
  testGetDescribeLocalVolume
  stopTest
}

@test "Deploy application" {
  startTest
  initApplicationFiles
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Deploy microservice" {
  startTest
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
  stopTest
}

@test "Deploy route" {
  startTest
  initRouteFile
  iofogctl -v -n "$NS" deploy -f test/conf/route.yaml
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Update microservice" {
  startTest
  initMicroserviceUpdateFile
  iofogctl -v -n "$NS" deploy -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Rename and Delete route" {
  startTest
  local NEW_ROUTE_NAME="route-2"
  iofogctl -v -n "$NS" rename route "$ROUTE_NAME" "$NEW_ROUTE_NAME"
  iofogctl -v -n "$NS" delete route "$NEW_ROUTE_NAME"
  checkRouteNegative "$NEW_ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  checkRouteNegative "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Delete microservice using file option" {
  startTest
  iofogctl -v -n "$NS" delete -f test/conf/updatedMicroservice.yaml
  checkMicroserviceNegative
  stopTest
}

@test "Deploy microservice in application" {
  startTest
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
  stopTest
}

@test "Get local logs" {
  startTest
  [[ ! -z $(iofogctl -v -n "$NS" logs controller $NAME) ]]
  [[ ! -z $(iofogctl -v -n "$NS" logs agent ${NAME}-0) ]]
  stopTest
}

@test "Deploy registry" {
  startTest
  initGCRRegistryFile
  iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkGCRRegistry
  initUpdatedGCRRegistryFile
  iofogctl -v -n "$NS" deploy -f test/conf/gcr.yaml
  checkUpdatedGCRRegistry
  iofogctl -v -n "$NS" delete registry 3
  checkGCRRegistryNegative
  stopTest
}

@test "Detach should fail because of running msvc" {
  startTest
  run iofogctl -v -n "$NS" detach agent ${NAME}-0
  [ "$status" -eq 1 ]
  echo "$output" | grep "because it still has microservices running. Remove the microservices first, or use the --force option."
  stopTest
}

@test "Detach/attach agent" {
  startTest
  iofogctl -v -n "$NS" detach agent ${NAME}-0 --force
  iofogctl -v -n "$NS" attach agent ${NAME}-0
  stopTest
}

@test "Deploy application from file and test application update" {
  startTest
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
  stopTest
}

@test "Delete agent should fail because of running msvc" {
  startTest
  run iofogctl -v -n "$NS" delete agent ${NAME}-0
  [ "$status" -eq 1 ]
  echo "$output" | grep "because it still has microservices running. Remove the microservices first, or use the --force option."
  stopTest
}

@test "Delete agent should work with --force option" {
  startTest
  iofogctl -v -n "$NS" delete agent ${NAME}-0 --force
  stopTest
}

@test "Delete all using file" {
  startTest
  initAllLocalDeleteFile
  iofogctl -v -n "$NS" delete -f test/conf/all-local.yaml
  checkApplicationNegative
  checkControllerNegative
  checkAgentNegative "${NAME}-0"
  stopTest
}

@test "Delete namespace" {
  startTest
  iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
  stopTest
}
