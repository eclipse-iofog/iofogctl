package deploylocalcontroller

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestNewExecutorRequiresLocalControlPlaneInNamespace(t *testing.T) {
	config.Init(t.TempDir())

	_, err := NewExecutor(Options{
		Namespace: "default",
		Yaml:      []byte("container:\n  image: ghcr.io/example/controller:1.0\n"),
		Name:      "iofog",
	})
	require.Error(t, err)
}

func TestNewExecutorAcceptsExistingLocalControlPlane(t *testing.T) {
	config.Init(t.TempDir())

	ns, err := config.GetNamespace("default")
	require.NoError(t, err)

	arch := "amd64"
	cp := &rsc.LocalControlPlane{
		IofogUser: rsc.IofogUser{Email: "user@domain.com"},
		Auth: rsc.Auth{
			Mode: "embedded",
			Bootstrap: &rsc.AuthBootstrap{
				Username: "admin",
				Password: "LocalTest12!",
			},
		},
		SystemAgent: &rsc.SystemAgentConfig{
			AgentConfiguration: &rsc.AgentConfiguration{Arch: &arch},
		},
	}
	require.NoError(t, cp.UpdateController(&rsc.LocalController{
		Name:     "iofog",
		Endpoint: "http://localhost:51121",
	}))
	ns.SetControlPlane(cp)
	require.NoError(t, config.Flush())

	exe, err := NewExecutor(Options{
		Namespace: "default",
		Yaml:      []byte("container:\n  image: ghcr.io/example/controller:1.0\n"),
		Name:      "iofog",
	})
	require.NoError(t, err)
	require.NotNil(t, exe)
	require.Equal(t, "iofog", exe.GetName())
}

func TestValidateRejectsInvalidControllerName(t *testing.T) {
	err := Validate(&rsc.LocalController{Name: "Invalid_Name"})
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}
