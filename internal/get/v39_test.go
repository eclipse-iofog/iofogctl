package get

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/stretchr/testify/require"
)

func TestGenerateModelOutput(t *testing.T) {
	table := generateModelOutput([]client.Model{
		{Name: "smollm2-135m", Repo: "ai/smollm2", RegistryID: 1, Format: "gguf", Revision: "sha256:abc"},
	})
	require.Equal(t, []string{"NAME", "REPO", "REGISTRY_ID", "FORMAT", "REVISION"}, table[0])
	require.Equal(t, []string{"smollm2-135m", "ai/smollm2", "1", "gguf", "sha256:abc"}, table[1])
}

func TestGenerateModelOutputEmpty(t *testing.T) {
	table := generateModelOutput(nil)
	require.Equal(t, []string{"NAME", "REPO", "REGISTRY_ID", "FORMAT", "REVISION"}, table[0])
	require.Len(t, table, 1)
}

func TestGenerateRuntimeClassOutput(t *testing.T) {
	table := generateRuntimeClassOutput([]client.RuntimeClass{
		{Name: "spin", Handler: "spin"},
	})
	require.Equal(t, []string{"NAME", "HANDLER"}, table[0])
	require.Equal(t, []string{"spin", "spin"}, table[1])
}

func TestGenerateMicroserviceTemplateOutput(t *testing.T) {
	table := generateMicroserviceTemplateOutput([]client.MicroserviceTemplate{
		{Name: "ms-template", Description: "Template for creating microservices", Variables: []client.MicroserviceTemplateVariable{{Key: "application"}, {Key: "agent-name"}}},
	})
	require.Equal(t, []string{"NAME", "DESCRIPTION", "VARIABLES"}, table[0])
	require.Equal(t, []string{"ms-template", "Template for creating microservices", "2"}, table[1])
}

func TestNewExecutorV39Resources(t *testing.T) {
	for _, resource := range []string{"models", "runtimeclass", "microservice-templates"} {
		exe, err := NewExecutor(resource, "default", false, "")
		require.NoError(t, err, resource)
		require.NotNil(t, exe, resource)
	}
}
