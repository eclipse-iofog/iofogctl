package deployk8scontrolplane

import (
	"testing"

	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

func requireInputError(t *testing.T, err error) {
	t.Helper()
	var inputErr *util.InputError
	require.ErrorAs(t, err, &inputErr)
}

func validControlPlane() *rsc.KubernetesControlPlane {
	return &rsc.KubernetesControlPlane{
		IofogUser: rsc.IofogUser{Email: "user@domain.com"},
		Auth: rsc.Auth{
			Mode: authModeEmbedded,
			Bootstrap: &rsc.AuthBootstrap{
				Username: "admin",
				Password: "LocalTest12!",
			},
		},
		Services: rsc.Services{
			Controller: rsc.Service{Type: "LoadBalancer"},
			Router:     rsc.Service{Type: "LoadBalancer"},
		},
		Replicas: rsc.Replicas{Controller: 1, Nats: 2},
	}
}

func TestValidateEmbeddedAuthMissingBootstrap(t *testing.T) {
	cp := validControlPlane()
	cp.Auth.Bootstrap = nil
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateEmbeddedAuthWeakBootstrapPassword(t *testing.T) {
	cp := validControlPlane()
	cp.Auth.Bootstrap.Password = "short"
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateExternalAuthMissingIssuerUrl(t *testing.T) {
	cp := validControlPlane()
	cp.Auth = rsc.Auth{
		Mode: authModeExternal,
		Client: &rsc.AuthClient{
			ID:     "controller",
			Secret: "secret-value",
		},
	}
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateExternalAuthMissingClientSecret(t *testing.T) {
	cp := validControlPlane()
	cp.Auth = rsc.Auth{
		Mode:      authModeExternal,
		IssuerUrl: "https://auth.example.com/realms/myrealm",
		Client: &rsc.AuthClient{
			ID: "controller",
		},
	}
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateExternalAuthValid(t *testing.T) {
	cp := validControlPlane()
	cp.Auth = rsc.Auth{
		Mode:      authModeExternal,
		IssuerUrl: "https://auth.example.com/realms/myrealm",
		Client: &rsc.AuthClient{
			ID:     "controller",
			Secret: "secret-value",
		},
	}
	require.NoError(t, validate(cp))
}

func TestValidateIofogUserWeakPassword(t *testing.T) {
	cp := validControlPlane()
	cp.IofogUser.Password = "weak"
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateControllerClusterIPWithoutIngress(t *testing.T) {
	cp := validControlPlane()
	cp.Services.Controller.Type = clusterIP
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateRouterClusterIPWithoutIngress(t *testing.T) {
	cp := validControlPlane()
	cp.Services.Router.Type = clusterIP
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateNatsReplicasOneWhenEnabled(t *testing.T) {
	cp := validControlPlane()
	cp.Replicas.Nats = 1
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateNatsReplicasSkippedWhenDisabled(t *testing.T) {
	cp := validControlPlane()
	disabled := false
	cp.Nats = &rsc.NatsSpec{Enabled: &disabled}
	cp.Replicas.Nats = 1
	require.NoError(t, validate(cp))
}

func TestValidateDatabaseHARequiresExternalDB(t *testing.T) {
	cp := validControlPlane()
	cp.Replicas.Controller = 2
	err := validate(cp)
	requireInputError(t, err)
}

func TestValidateValidEmbeddedControlPlane(t *testing.T) {
	require.NoError(t, validate(validControlPlane()))
}
