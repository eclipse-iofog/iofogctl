package deploymicroservice

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
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

func TestUnmarshalMicroserviceSupplementalGroups(t *testing.T) {
	var missing apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict([]byte("application: test-app\ncontainer:\n  runAsGroup: \"1000\"\n"), &missing))
	require.Nil(t, missing.Container.SupplementalGroups)

	var cleared apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict([]byte("application: test-app\ncontainer:\n  supplementalGroups: []\n"), &cleared))
	require.NotNil(t, cleared.Container.SupplementalGroups)
	require.Empty(t, *cleared.Container.SupplementalGroups)

	var microservice apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict([]byte("application: test-app\ncontainer:\n  supplementalGroups:\n    - 993\n    - 44\n"), &microservice))
	require.NotNil(t, microservice.Container.SupplementalGroups)
	require.Equal(t, []int{993, 44}, *microservice.Container.SupplementalGroups)
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

func TestUnmarshalMicroserviceSpecKeepsModelsAndKnowledge(t *testing.T) {
	raw := []byte(`
application: my-app
models:
  bindPath: /models
  permissions: ro
  items:
    - name: smollm2-135m
knowledge:
  bindPath: /knowledge
  permissions: ro
  items:
    - name: product-docs
`)
	var microservice apps.Microservice
	require.NoError(t, yaml.UnmarshalStrict(raw, &microservice))
	require.NotNil(t, microservice.Models)
	require.Equal(t, "/models", microservice.Models.BindPath)
	require.Equal(t, "smollm2-135m", microservice.Models.Items[0].Name)
	require.NotNil(t, microservice.Knowledge)
	require.Equal(t, "/knowledge", microservice.Knowledge.BindPath)
	require.Equal(t, "ro", microservice.Knowledge.Permissions)
	require.Equal(t, "product-docs", microservice.Knowledge.Items[0].Name)
}

func TestParseMicroserviceKnowledge(t *testing.T) {
	raw := []byte(`
application: my-app
models:
  bindPath: /models
  permissions: ro
  items:
    - name: smollm2-135m
knowledge:
  bindPath: /knowledge
  permissions: ro
  items:
    - name: product-docs
`)
	catalog, err := parseMicroserviceKnowledge(raw)
	require.NoError(t, err)
	require.NotNil(t, catalog)
	require.Equal(t, "/knowledge", catalog.BindPath)
	require.Equal(t, "ro", catalog.Permissions)
	require.Len(t, catalog.Items, 1)
	require.Equal(t, "product-docs", catalog.Items[0].Name)

	clientCatalog := toClientKnowledgeCatalog(*catalog)
	require.Equal(t, "/knowledge", clientCatalog.BindPath)
	require.Equal(t, "ro", clientCatalog.Permissions)
	require.Equal(t, "product-docs", clientCatalog.Items[0].Name)
}

func TestParseMicroserviceKnowledgeMissing(t *testing.T) {
	catalog, err := parseMicroserviceKnowledge([]byte("application: test-app\n"))
	require.NoError(t, err)
	require.Nil(t, catalog)
}

func TestParseMicroserviceKnowledgeEmptyObjectPresent(t *testing.T) {
	catalog, err := parseMicroserviceKnowledge([]byte("knowledge: {}\n"))
	require.NoError(t, err)
	require.NotNil(t, catalog)
	require.Empty(t, catalog.BindPath)
	require.Empty(t, catalog.Items)
}

func TestReconcileMicroserviceIdentity(t *testing.T) {
	cases := []struct {
		name     string
		metadata string
		spec     string
		wantApp  string
		wantMsvc string
		wantErr  string
	}{
		{
			name:     "fq metadata and application",
			metadata: "inference/pm-client",
			spec:     "application: inference\n",
			wantApp:  "inference",
			wantMsvc: "pm-client",
		},
		{
			name:     "short metadata and application",
			metadata: "pm-client",
			spec:     "application: inference\n",
			wantApp:  "inference",
			wantMsvc: "pm-client",
		},
		{
			name:     "fq metadata explicit spec name",
			metadata: "inference/pm-client",
			spec:     "name: pm-client\napplication: inference\n",
			wantApp:  "inference",
			wantMsvc: "pm-client",
		},
		{
			name:     "short metadata explicit spec name",
			metadata: "pm-client",
			spec:     "name: pm-client\napplication: inference\n",
			wantApp:  "inference",
			wantMsvc: "pm-client",
		},
		{
			name:     "fq metadata without application",
			metadata: "inference/pm-client",
			spec:     "agent:\n  name: edge-01\n",
			wantApp:  "inference",
			wantMsvc: "pm-client",
		},
		{
			name:     "application mismatch",
			metadata: "inference/pm-client",
			spec:     "application: other-app\n",
			wantErr:  "does not match spec.application",
		},
		{
			name:     "spec name mismatch",
			metadata: "inference/pm-client",
			spec:     "name: other-svc\napplication: inference\n",
			wantErr:  "does not match metadata.name",
		},
		{
			name:     "spec name is fully qualified",
			metadata: "pm-client",
			spec:     "name: inference/pm-client\napplication: inference\n",
			wantErr:  "must be the microservice name",
		},
		{
			name:     "missing application",
			metadata: "pm-client",
			spec:     "agent:\n  name: edge-01\n",
			wantErr:  "Application name missing",
		},
		{
			name:     "invalid metadata name",
			metadata: "Inference/pm-client",
			spec:     "application: inference\n",
			wantErr:  "not valid",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var microservice apps.Microservice
			require.NoError(t, yaml.UnmarshalStrict([]byte(tc.spec), &microservice))
			appName, msvcName, err := reconcileMicroserviceIdentity(tc.metadata, &microservice)
			if tc.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantApp, appName)
			require.Equal(t, tc.wantMsvc, msvcName)
			require.Equal(t, tc.wantMsvc, microservice.Name)
			require.Equal(t, tc.wantApp, microservice.Application)
		})
	}
}

func TestDeployMicroserviceYAMLIncludesSpecName(t *testing.T) {
	cases := []struct {
		name     string
		metadata string
		spec     string
	}{
		{
			name:     "fq metadata",
			metadata: "inference/pm-client",
			spec:     "application: inference\nagent:\n  name: edge-01\n",
		},
		{
			name:     "short metadata",
			metadata: "pm-client",
			spec:     "application: inference\nagent:\n  name: edge-01\n",
		},
		{
			name:     "fq metadata without application",
			metadata: "inference/pm-client",
			spec:     "agent:\n  name: edge-01\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var microservice apps.Microservice
			require.NoError(t, yaml.UnmarshalStrict([]byte(tc.spec), &microservice))
			appName, msvcName, err := reconcileMicroserviceIdentity(tc.metadata, &microservice)
			require.NoError(t, err)

			for _, update := range []bool{false, true} {
				path := "create"
				if update {
					path = "update"
				}
				t.Run(path, func(t *testing.T) {
					raw := deployMicroserviceYAML(t, microservice, appName, msvcName, update)
					requireMicroserviceSpecName(t, raw, appName, msvcName)
				})
			}
		})
	}
}

func deployMicroserviceYAML(t *testing.T, microservice apps.Microservice, appName, msvcName string, update bool) []byte {
	t.Helper()
	const uuid = "10674117-42f3-473e-8e50-aa700d55ae46"
	var captured []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/status":
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/microservices":
			if update {
				_, _ = w.Write([]byte(`{"microservices":[{"uuid":"` + uuid + `","name":"` + msvcName + `"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"microservices":[]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v3/microservices/yaml":
			captured = microserviceYAMLFromRequest(t, r)
			_, _ = w.Write([]byte(`{"uuid":"` + uuid + `"}`))
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v3/microservices/yaml/"+uuid:
			captured = microserviceYAMLFromRequest(t, r)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v3/microservices/"+uuid:
			_, _ = w.Write([]byte(`{"uuid":"` + uuid + `","name":"` + msvcName + `"}`))
		default:
			http.Error(w, "unexpected "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	err := apps.DeployMicroservice(apps.IofogController{
		Endpoint: srv.URL,
		Token:    "test-token",
	}, &microservice, appName, msvcName, apps.WithAPIVersion(util.GetCliApiVersion()))
	require.NoError(t, err)
	require.NotEmpty(t, captured)
	return captured
}

func microserviceYAMLFromRequest(t *testing.T, r *http.Request) []byte {
	t.Helper()
	require.NoError(t, r.ParseMultipartForm(1<<20))
	file, _, err := r.FormFile("microservice")
	require.NoError(t, err)
	defer file.Close()
	body, err := io.ReadAll(file)
	require.NoError(t, err)
	return body
}

type deployedMicroserviceMetadata struct {
	Name string `yaml:"name"`
}

type deployedMicroserviceSpec struct {
	Name        string `yaml:"name"`
	Application string `yaml:"application"`
}

type deployedMicroserviceDoc struct {
	Metadata deployedMicroserviceMetadata `yaml:"metadata"`
	Spec     deployedMicroserviceSpec     `yaml:"spec"`
}

func requireMicroserviceSpecName(t *testing.T, raw []byte, appName, msvcName string) {
	t.Helper()
	var doc deployedMicroserviceDoc
	require.NoError(t, yaml.Unmarshal(raw, &doc))
	require.Equal(t, appName+"/"+msvcName, doc.Metadata.Name)
	require.Equal(t, msvcName, doc.Spec.Name)
	require.Equal(t, appName, doc.Spec.Application)
}
