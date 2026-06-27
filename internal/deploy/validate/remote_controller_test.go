package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeDeployFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "deploy.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

func TestRemoteControllerDeployNoCollision(t *testing.T) {
	path := writeDeployFixture(t, `apiVersion: datasance.com/v3
kind: ControlPlane
metadata:
  name: cp
spec:
  controllers:
    - name: remote-1
      host: 10.0.0.1
---
apiVersion: datasance.com/v3
kind: Controller
metadata:
  name: remote-2
spec:
  host: 10.0.0.2
`)
	require.NoError(t, RemoteControllerDeploy(path))
}

func TestRemoteControllerDeployNameCollision(t *testing.T) {
	path := writeDeployFixture(t, `apiVersion: datasance.com/v3
kind: ControlPlane
metadata:
  name: cp
spec:
  controllers:
    - name: remote-1
      host: 10.0.0.1
---
apiVersion: datasance.com/v3
kind: Controller
metadata:
  name: remote-1
spec:
  host: 10.0.0.2
`)
	err := RemoteControllerDeploy(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "name")
}

func TestRemoteControllerDeployControlPlaneOnlyFullSpec(t *testing.T) {
	path := writeDeployFixture(t, `apiVersion: datasance.com/v3
kind: ControlPlane
metadata:
  name: iofog
spec:
  iofogUser:
    name: Foo
    email: user@domain.com
    password: "TestPassword12!"
  controller:
    publicUrl: http://192.168.139.85:51121
    consoleUrl: http://192.168.139.85
  auth:
    mode: embedded
    bootstrap:
      username: admin
      password: "LocalTest12!"
  systemMicroservices:
    router:
      amd64: ghcr.io/datasance/router:3.8.0-rc.1
  nats:
    enabled: true
  controllers:
    - name: remote-1
      host: 0.0.0.0
      ssh:
        user: ubuntu
        keyFile: ~/.ssh/id_ed25519
        port: 32222
      systemAgent:
        config:
          host: 192.168.139.85
          arch: arm64
          containerEngine: edgelet
          deploymentType: native
`)
	require.NoError(t, RemoteControllerDeploy(path))
}

func TestRemoteControllerDeployHostCollision(t *testing.T) {
	path := writeDeployFixture(t, `apiVersion: datasance.com/v3
kind: ControlPlane
metadata:
  name: cp
spec:
  controllers:
    - name: remote-1
      host: 10.0.0.1
---
apiVersion: datasance.com/v3
kind: RemoteController
metadata:
  name: remote-2
spec:
  host: 10.0.0.1
`)
	err := RemoteControllerDeploy(path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "host")
}
