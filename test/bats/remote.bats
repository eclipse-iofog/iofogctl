#!/usr/bin/env bash

. test/func/include.bash

NS="$NAMESPACE"

@test "Initialize tests" {
  stopTest
}

@test "Verify remote controller host configured" {
  startTest
  initVanillaController
  [[ ! -z "$VANILLA_CONTROLLER" && "$VANILLA_CONTROLLER" != "user@host" ]]
  stopTest
}

@test "Create namespace" {
  startTest
  iofogctl create namespace "$NS"
  stopTest
}

@test "Deploy remote ControlPlane" {
  startTest
  initRemoteControlPlaneFile
  iofogctl -v -n "$NS" deploy -f test/conf/remote-cp.yaml
  checkRemoteControlPlane
  stopTest
}

@test "Describe remote controller" {
  startTest
  checkRemoteControlPlane
  stopTest
}

@test "Connect remote ControlPlane from file" {
  startTest
  iofogctl -v -n "$NS" connect -f test/conf/remote-cp.yaml
  checkRemoteControlPlane
  stopTest
}

@test "Deploy add-on remote Controller" {
  startTest
  initRemoteControllerAddFile
  iofogctl -v -n "$NS" deploy -f test/conf/remote-controller-add.yaml
  [[ $(iofogctl -v -n "$NS" get controllers | grep -c "$NAME") -ge 1 ]]
  stopTest
}

@test "Delete one remote Controller" {
  startTest
  initRemoteControllerAddFile
  local REMOTE_NAME2="${REMOTE_CONTROLLER2_NAME:-${NAME}-2}"
  iofogctl -v -n "$NS" delete controller "$REMOTE_NAME2"
  [[ -z $(iofogctl -v -n "$NS" get controllers | grep "$REMOTE_NAME2") ]]
  [[ ! -z $(iofogctl -v -n "$NS" get controllers | grep "$NAME") ]]
  initVanillaController
  checkRemoteEdgeletDeleted "$REMOTE_CONTROLLER2_USER" "$REMOTE_CONTROLLER2_HOST" "${REMOTE_CONTROLLER2_PORT:-22}" "$KEY_FILE"
  stopTest
}

@test "Delete remote ControlPlane" {
  startTest
  initVanillaController
  iofogctl -v -n "$NS" delete controlplane
  checkControllerNegative
  checkRemoteEdgeletDeleted "$VANILLA_USER" "$VANILLA_HOST" "$VANILLA_PORT" "$KEY_FILE"
  stopTest
}

@test "Delete namespace" {
  startTest
  iofogctl -v delete namespace "$NS"
  [[ -z $(iofogctl get namespaces | grep "$NS") ]]
  stopTest
}
