package deleteremotecontrolplane

import (
	"errors"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

type stubEdgeletHost struct {
	deprovisionErr   error
	deleteCPErr      error
	uninstallErr     error
	deprovisionCalls int
	deleteCPCalls    int
	uninstallCalls   int
}

func (s *stubEdgeletHost) Deprovision() error {
	s.deprovisionCalls++
	return s.deprovisionErr
}

func (s *stubEdgeletHost) DeleteControlPlane() error {
	s.deleteCPCalls++
	return s.deleteCPErr
}

func (s *stubEdgeletHost) Uninstall(bool) error {
	s.uninstallCalls++
	return s.uninstallErr
}

func TestIsIgnorableControllerAgentDeleteError(t *testing.T) {
	require.True(t, isIgnorableControllerAgentDeleteError(util.NewNotFoundError("agent")))
	require.True(t, isIgnorableControllerAgentDeleteError(errors.New("connection refused")))
	require.True(t, isIgnorableControllerAgentDeleteError(errors.New("connect: connection refused")))
	require.False(t, isIgnorableControllerAgentDeleteError(errors.New("permission denied")))
}

func TestResolveSystemAgentUUIDFromConfig(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	ns, err := config.GetNamespace("default")
	require.NoError(t, err)
	require.NoError(t, ns.AddAgent(&rsc.RemoteAgent{
		Name: "remote-1",
		UUID: "uuid-from-config",
	}))
	require.NoError(t, config.Flush())

	require.Equal(t, "uuid-from-config", resolveSystemAgentUUID("default", "remote-1"))
}

func TestTeardownRemoteControllerHostContinuesOnEdgeletErrors(t *testing.T) {
	t.Cleanup(resetTeardownHooks())

	stub := &stubEdgeletHost{
		deprovisionErr: errors.New("deprovision failed"),
		deleteCPErr:    errors.New("delete cp failed"),
		uninstallErr:   errors.New("uninstall failed"),
	}
	buildRemoteEdgeletFn = func(*rsc.RemoteControlPlane, *rsc.RemoteController, string) (*install.RemoteEdgelet, error) {
		return &install.RemoteEdgelet{}, nil
	}
	edgeletDeprovisionFn = func(edgeletHostTeardown) error { return stub.Deprovision() }
	edgeletDeleteCPFn = func(edgeletHostTeardown) error { return stub.DeleteControlPlane() }
	edgeletUninstallFn = func(edgeletHostTeardown) error { return stub.Uninstall(true) }
	deleteSystemAgentFn = func(string, string, string) {}

	cp := &rsc.RemoteControlPlane{}
	ctrl := &rsc.RemoteController{Name: "remote-1", Host: "10.0.0.1"}

	require.NoError(t, TeardownRemoteControllerHost("default", cp, ctrl))
	require.Equal(t, 1, stub.deprovisionCalls)
	require.Equal(t, 1, stub.deleteCPCalls)
	require.Equal(t, 1, stub.uninstallCalls)
}

func TestExecutorExecuteRemovesCA(t *testing.T) {
	t.Cleanup(resetTeardownHooks())

	dir := t.TempDir()
	config.Init(dir)

	arch := "amd64"
	cp := &rsc.RemoteControlPlane{
		Controllers: []rsc.RemoteController{{
			Name: "remote-1",
			Host: "10.0.0.1",
			SystemAgent: &rsc.SystemAgentConfig{
				AgentConfiguration: &rsc.AgentConfiguration{Arch: &arch},
			},
		}},
	}
	ns, err := config.GetNamespace("default")
	require.NoError(t, err)
	ns.SetControlPlane(cp)
	require.NoError(t, config.Flush())
	require.NoError(t, trust.StoreCA("default", "dGVzdA=="))

	buildRemoteEdgeletFn = func(*rsc.RemoteControlPlane, *rsc.RemoteController, string) (*install.RemoteEdgelet, error) {
		return &install.RemoteEdgelet{}, nil
	}
	edgeletDeprovisionFn = func(edgeletHostTeardown) error { return nil }
	edgeletDeleteCPFn = func(edgeletHostTeardown) error { return nil }
	edgeletUninstallFn = func(edgeletHostTeardown) error { return nil }
	deleteSystemAgentFn = func(string, string, string) {}

	exe, err := NewExecutor("default")
	require.NoError(t, err)
	require.NoError(t, exe.Execute())
	require.False(t, trust.HasCA("default"))

	ns, err = config.GetNamespace("default")
	require.NoError(t, err)
	_, err = ns.GetControlPlane()
	require.Error(t, err)
}

func resetTeardownHooks() func() {
	prevBuild := buildRemoteEdgeletFn
	prevDeleteAgent := deleteSystemAgentFn
	prevDeprovision := edgeletDeprovisionFn
	prevDeleteCP := edgeletDeleteCPFn
	prevUninstall := edgeletUninstallFn
	return func() {
		buildRemoteEdgeletFn = prevBuild
		deleteSystemAgentFn = prevDeleteAgent
		edgeletDeprovisionFn = prevDeprovision
		edgeletDeleteCPFn = prevDeleteCP
		edgeletUninstallFn = prevUninstall
	}
}
