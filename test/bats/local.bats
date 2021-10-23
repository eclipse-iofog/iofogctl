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

@test "Test wrong namespace metadata" {
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

@test "Deploy Application for docker pull stats" {
  startTest
  initDockerPullStatsApplicationFiles
  iofogctl -v -n "$NS" deploy -f test/conf/application_pull_stat.yaml
  waitForPullingMsvc "$MSVC5_NAME" "$NS"
  checkPullPercentageOfMicroservice "$MSVC5_NAME" "$NS"
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

@test "Agent config dev mode" {
  startTest  
  [[ ! -z $(iofogctl -v -n "$NS" legacy agent "${NAME}-0" 'config -dev on') ]]
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

@test "Deploy Application Template and Templated Application" {
  startTest
  testApplicationTemplates
  stopTest
}

@test "Deploy Application" {
  startTest
  initApplicationFiles
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Deploy Microservice" {
  startTest
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
  stopTest
}

@test "Deploy Route" {
  startTest
  initRouteFile
  iofogctl -v -n "$NS" deploy -f test/conf/route.yaml
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Update Microservice" {
  startTest
  initMicroserviceUpdateFile
  iofogctl --debug -n "$NS" deploy -f test/conf/updatedMicroservice.yaml
  checkUpdatedMicroservice
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Rename and Delete Route" {
  startTest
  local NEW_ROUTE_NAME="route-2"
  iofogctl -v -n "$NS" rename route $APPLICATION_NAME/"$ROUTE_NAME" "$NEW_ROUTE_NAME"
  iofogctl -v -n "$NS" delete route $APPLICATION_NAME/"$NEW_ROUTE_NAME"
  checkRouteNegative "$NEW_ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  checkRouteNegative "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}

@test "Delete Microservice using file option" {
  startTest
  iofogctl -v -n "$NS" delete -f test/conf/updatedMicroservice.yaml
  checkMicroserviceNegative
  stopTest
}

@test "Deploy Microservice in Application" {
  startTest
  initMicroserviceFile
  iofogctl -v -n "$NS" deploy -f test/conf/microservice.yaml
  checkMicroservice
  stopTest
}

@test "Get local json logs" {
  startTest
  [[ ! -z $(iofogctl -v -n "$NS" logs controller $NAME) ]]
  [[ ! -z $(iofogctl -v -n "$NS" logs agent ${NAME}-0) ]]
  [iofogctl logs ${NAME}-0 -n "$NS" | jq -e . >/dev/null 2>&1  | echo ${PIPESTATUS[1]}  -eq 0 ]
  stopTest
}

@test "Deploy Registry" {
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

@test "Detach/attach Agent" {
  startTest
  iofogctl -v -n "$NS" detach agent ${NAME}-0 --force
  iofogctl -v describe agent ${NAME}-0 --detached
  iofogctl -v -n "$NS" attach agent ${NAME}-0
  stopTest
}

@test "Deploy Application from file and test Application update" {
  startTest
  iofogctl -v -n "$NS" deploy -f test/conf/application.yaml
  checkApplication
  stopTest
}

@test "Deploy Application with volume missing " {
  startTest
  initInvalidApplicationFiles
  iofogctl -v -n "$NS" deploy -f test/conf/application_volume_missing.yaml
  waitForFailedMsvc "$MSVC4_NAME" "$NS"
  [[ ! -z $(iofogctl get microservices -n "$NS"  | grep "Volume missing") ]]
  iofogctl get microservices -n "$NS"
  stopTest
}

@test "Delete Agent should fail because of running msvc" {
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

@test "Delete Namespace" {
  startTest
  iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
  stopTest
}
