package deploymicroservice

import (
	"testing"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalMicroserviceSpecKeepsTemplateAndModels(t *testing.T) {
	raw := []byte(`
template:
  name: ms-template
  variables:
    - key: agent-name
      value: lima
    - key: model1
      value: tiny-gpt2
    - key: application
      value: test-app
models:
  bindPath: /models
  permissions: ro
  items:
    - name: tiny-gpt2
application: test-app
`)
	var microservice apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict(raw, &microservice))
	require.NotNil(t, microservice.Template)
	require.Equal(t, "ms-template", microservice.Template.Name)
	require.NotNil(t, microservice.Models)
	require.Equal(t, "/models", microservice.Models.BindPath)
}

func TestUnmarshalMicroserviceSpecKeepsServiceAccount(t *testing.T) {
	raw := []byte(`
images:
  registry: 1
  amd64: ghcr.io/example/app:latest
serviceAccount:
  roleRef:
    kind: Role
    name: microservice
application: test-app
`)
	var microservice apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict(raw, &microservice))
	require.NotNil(t, microservice.ServiceAccount)
	require.Equal(t, "Role", microservice.ServiceAccount.RoleRef.Kind)
	require.Equal(t, "microservice", microservice.ServiceAccount.RoleRef.Name)
	require.Equal(t, apps.RegistryRef(1), microservice.Images.Registry)
}

func TestParseMicroserviceModels(t *testing.T) {
	raw := []byte(`
application: test-app
models:
  bindPath: /models
  permissions: ro
  items:
    - name: tiny-gpt2
`)
	catalog, err := parseMicroserviceModels(raw)
	require.NoError(t, err)
	require.NotNil(t, catalog)
	require.Equal(t, "/models", catalog.BindPath)
	require.Equal(t, "ro", catalog.Permissions)
	require.Len(t, catalog.Items, 1)
	require.Equal(t, "tiny-gpt2", catalog.Items[0].Name)

	clientCatalog := toClientCatalog(*catalog)
	require.Equal(t, "/models", clientCatalog.BindPath)
	require.Equal(t, "ro", clientCatalog.Permissions)
	require.Equal(t, "tiny-gpt2", clientCatalog.Items[0].Name)
}

func TestParseMicroserviceModelsMissing(t *testing.T) {
	catalog, err := parseMicroserviceModels([]byte("application: test-app\n"))
	require.NoError(t, err)
	require.Nil(t, catalog)
}

func TestParseMicroserviceModelsEmptyObjectPresent(t *testing.T) {
	catalog, err := parseMicroserviceModels([]byte("models: {}\n"))
	require.NoError(t, err)
	require.NotNil(t, catalog)
	require.Empty(t, catalog.BindPath)
	require.Empty(t, catalog.Items)
}

func TestResolveMicroserviceNameFromApplication(t *testing.T) {
	require.Equal(t, "test-app/tiny-gpt2-ms", resolveMicroserviceName("tiny-gpt2-ms", []byte("application: test-app\n")))
	require.Equal(t, "other/ms", resolveMicroserviceName("other/ms", []byte("application: test-app\n")))
	require.Equal(t, "tiny-gpt2-ms", resolveMicroserviceName("tiny-gpt2-ms", []byte("template:\n  name: ms-template\n")))
}
