package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func fullRemoteControlPlane() *RemoteControlPlane {
	return &RemoteControlPlane{
		Auth: Auth{Mode: "embedded"},
		Controller: LocalControllerSpec{
			ControllerConfig: ControllerConfig{PublicUrl: "http://192.168.1.6:51121"},
		},
		Controllers: []RemoteController{{Name: "remote-1", Host: "10.0.0.1"}},
	}
}

func TestSupportsControllerAddOnThinCPRejected(t *testing.T) {
	cp := &RemoteControlPlane{
		IofogUser:   IofogUser{Email: "user@example.com"},
		Controllers: []RemoteController{{Name: "ctrl-1", Host: "10.0.0.1", Endpoint: "http://10.0.0.1:51121"}},
	}
	err := cp.SupportsControllerAddOn()
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect -f")
}

func TestSupportsControllerAddOnMissingControllerBlock(t *testing.T) {
	cp := &RemoteControlPlane{
		Auth:        Auth{Mode: "embedded"},
		Controllers: []RemoteController{{Name: "remote-1", Host: "10.0.0.1"}},
	}
	err := cp.SupportsControllerAddOn()
	require.Error(t, err)
	require.Contains(t, err.Error(), "spec.controller")
}

func TestSupportsControllerAddOnFullCPOK(t *testing.T) {
	cp := fullRemoteControlPlane()
	require.NoError(t, cp.SupportsControllerAddOn())
}

func TestValidateControllerAddOnDatabaseSQLiteRejected(t *testing.T) {
	cp := fullRemoteControlPlane()
	err := cp.ValidateControllerAddOnDatabase()
	require.Error(t, err)
	require.Contains(t, err.Error(), "external database")
}

func TestValidateControllerAddOnDatabaseExternalOK(t *testing.T) {
	cp := fullRemoteControlPlane()
	cp.Database = Database{Provider: "postgres", Host: "db.example.com"}
	require.NoError(t, cp.ValidateControllerAddOnDatabase())
}

func TestValidateControllerAddOnDuplicateName(t *testing.T) {
	cp := fullRemoteControlPlane()
	err := cp.ValidateControllerAddOn(&RemoteController{Name: "remote-1", Host: "10.0.0.2"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "name")
}

func TestValidateControllerAddOnDuplicateHost(t *testing.T) {
	cp := fullRemoteControlPlane()
	err := cp.ValidateControllerAddOn(&RemoteController{Name: "remote-2", Host: "10.0.0.1"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "host")
}

func TestValidateControllerAddOnUniqueOK(t *testing.T) {
	cp := fullRemoteControlPlane()
	err := cp.ValidateControllerAddOn(&RemoteController{Name: "remote-2", Host: "10.0.0.2"})
	require.NoError(t, err)
}
