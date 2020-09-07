@test "Deploy Volumes" {
  startTest
  testDeployVolume
  testGetDescribeVolume
  stopTest
}

@test "Set default namespace" {
  startTest
  testDefaultNamespace "$NS"
  stopTest
}

@test "Deploy application" {
  startTest
  initApplicationFiles
  iofogctl -v deploy -f test/conf/application.yaml
  checkApplication
  waitForMsvc "$MSVC1_NAME" "$NS"
  waitForMsvc "$MSVC2_NAME" "$NS"
  stopTest
}

@test "Deploy route" {
  startTest
  initRouteFile
  iofogctl -v -n "$NS" deploy -f test/conf/route.yaml
  checkRoute "$ROUTE_NAME" "$MSVC1_NAME" "$MSVC2_NAME"
  stopTest
}
