package describe

import (
	"strings"
	"testing"
	"time"

	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestModelHeaderYAML(t *testing.T) {
	header := modelHeader("default", "smollm2-135m", &client.Model{
		UUID:       "model-uuid",
		Name:       "smollm2-135m",
		Repo:       "ai/smollm2",
		Revision:   "sha256:abc",
		RegistryID: 1,
		Format:     "gguf",
	}, []string{"lima"})

	require.Equal(t, config.ModelKind, header.Kind)
	out, err := yaml.Marshal(header)
	require.NoError(t, err)
	text := string(out)
	require.Contains(t, text, "kind: Model")
	require.Contains(t, text, "uuid: model-uuid")
	require.Contains(t, text, "registryId: 1")
	require.Contains(t, text, "linkedAgents:")
	require.Contains(t, text, "- lima")
	require.True(t, strings.Contains(text, "spec:"))
}

func TestRuntimeClassHeaderYAML(t *testing.T) {
	header := runtimeClassHeader("default", "spin", "spin", []string{"lima"})
	require.Equal(t, config.RuntimeClassKind, header.Kind)
	require.Equal(t, "spin", header.Handler)
	require.Nil(t, header.Spec)

	out, err := yaml.Marshal(header)
	require.NoError(t, err)
	text := string(out)
	require.Contains(t, text, "kind: RuntimeClass")
	require.Contains(t, text, "handler: spin")
	require.NotContains(t, text, "spec:")
	require.Contains(t, text, "linkedAgents:")
	require.Contains(t, text, "- lima")
	require.Less(t, strings.Index(text, "handler: spin"), strings.Index(text, "linkedAgents:"))
}

func TestMicroserviceTemplateHeaderYAML(t *testing.T) {
	header := microserviceTemplateHeader("default", "ms-template", &client.MicroserviceTemplate{
		Name:        "ms-template",
		Description: "Template for creating microservices",
		Variables: []client.MicroserviceTemplateVariable{
			{Key: "application", Description: "Application name", DefaultValue: "test-app"},
		},
		Microservice: map[string]any{"application": "{{application}}"},
	})
	require.Equal(t, config.MicroserviceTemplateKind, header.Kind)

	out, err := yaml.Marshal(header)
	require.NoError(t, err)
	text := string(out)
	require.Contains(t, text, "kind: MicroserviceTemplate")
	require.Contains(t, text, "description: Template for creating microservices")
	require.Contains(t, text, "key: application")
	require.Contains(t, text, "defaultValue: test-app")
}

func TestDescribeFactoryV39Resources(t *testing.T) {
	for _, resource := range []string{"model", "runtimeclass", "microservice-template"} {
		exe, err := NewExecutor(&Options{Resource: resource, Name: "demo", Namespace: "default"})
		require.NoError(t, err, resource)
		require.NotNil(t, exe, resource)
		require.Equal(t, "demo", exe.GetName())
	}
}

func TestConstructMicroserviceV39Fields(t *testing.T) {
	msvc, status, _, err := constructMicroservice(&client.MicroserviceInfo{
		Name:                   "infer",
		Config:                 "{}",
		Annotations:            "{}",
		RunAsGroup:             "1000",
		ReadOnlyRootFilesystem: true,
		Entrypoint:             []string{"/usr/bin/python"},
		WorkingDir:             "/app",
		Cpus:                   1.5,
		Commands:               []string{"-c"},
		CommandsAlias:          []string{"-m", "app"},
		Runtime:                "nvidia",
		CdiDevices:             []string{"nvidia.com/gpu=0"},
		Sysctls:                map[string]string{"net.core.somaxconn": "1024"},
		Ulimits:                map[string]client.ContainerUlimit{"nofile": {Soft: 1024, Hard: 2048}},
		MemoryReservation:      128,
		MemorySwap:             256,
		ShmSize:                64,
		Tmpfs:                  []client.ContainerTmpfs{{ContainerPath: "/tmp", Size: 64}},
		Devices:                []client.ContainerDevice{{HostPath: "/dev/nvidia0", ContainerPath: "/dev/nvidia0", Permissions: "rwm"}},
		Models: &client.MicroserviceCatalog{
			BindPath:    "/models",
			Permissions: "ro",
			Items:       []client.MicroserviceCatalogItem{{Name: "llama"}},
		},
		Status: client.MicroserviceStatusInfo{Status: "RUNNING", PodID: "pod-abc"},
	}, "lima", "test-app", nil)
	require.NoError(t, err)
	require.Equal(t, "1000", msvc.Container.RunAsGroup)
	require.True(t, msvc.Container.ReadOnlyRootFilesystem)
	require.Equal(t, []string{"/usr/bin/python"}, msvc.Container.Entrypoint)
	require.Equal(t, "/app", msvc.Container.WorkingDir)
	require.NotNil(t, msvc.Container.CPUs)
	require.Equal(t, 1.5, *msvc.Container.CPUs)
	require.Equal(t, []string{"-m", "app"}, msvc.Container.Commands)
	require.Equal(t, "nvidia", msvc.Container.Runtime)
	require.Equal(t, []string{"nvidia.com/gpu=0"}, msvc.Container.CdiDevices)
	require.Equal(t, "1024", msvc.Container.Sysctls["net.core.somaxconn"])
	require.Equal(t, 1024, msvc.Container.Ulimits["nofile"].Soft)
	require.NotNil(t, msvc.Container.MemoryReservation)
	require.Equal(t, int64(128), *msvc.Container.MemoryReservation)
	require.NotNil(t, msvc.Container.ShmSize)
	require.Equal(t, int64(64), *msvc.Container.ShmSize)
	require.Len(t, msvc.Container.Tmpfs, 1)
	require.Equal(t, "/tmp", msvc.Container.Tmpfs[0].ContainerPath)
	require.NotNil(t, msvc.Container.Tmpfs[0].Size)
	require.Equal(t, int64(64), *msvc.Container.Tmpfs[0].Size)
	require.Equal(t, "/dev/nvidia0", msvc.Container.Devices[0].HostPath)
	require.NotNil(t, msvc.Models)
	require.Equal(t, "/models", msvc.Models.BindPath)
	require.Equal(t, "llama", msvc.Models.Items[0].Name)
	require.Equal(t, "pod-abc", status.PodID)

	formatted := FormatMicroserviceStatus(status)
	require.Equal(t, "pod-abc", formatted["podId"])
	require.Equal(t, "", formatted["lastError"])
	require.Equal(t, int64(0), formatted["lastErrorAt"])
	require.Equal(t, 0, formatted["restartCount"])
}

func TestConstructMicroserviceStatusErrorExtras(t *testing.T) {
	const lastErrorAt int64 = 1710000000123
	_, status, _, err := constructMicroservice(&client.MicroserviceInfo{
		Name:        "infer",
		Config:      "{}",
		Annotations: "{}",
		Status: client.MicroserviceStatusInfo{
			Status:       "RUNNING",
			ErrorMessage: "",
			LastError:    "OOMKilled",
			LastErrorAt:  lastErrorAt,
			RestartCount: 2,
		},
	}, "lima", "test-app", nil)
	require.NoError(t, err)
	require.Equal(t, "OOMKilled", status.LastError)
	require.Equal(t, lastErrorAt, status.LastErrorAt)
	require.Equal(t, 2, status.RestartCount)

	formatted := FormatMicroserviceStatus(status)
	require.Equal(t, "OOMKilled", formatted["lastError"])
	require.Equal(t, 2, formatted["restartCount"])
	require.Equal(t, "", formatted["errorMessage"])
	wantAt := time.Unix(lastErrorAt/1000, (lastErrorAt%1000)*1000000).Format(time.RFC3339)
	require.Equal(t, wantAt, formatted["lastErrorAt"])
}

func TestConstructMicroserviceRegistryAndServiceAccount(t *testing.T) {
	msvc, _, _, err := constructMicroservice(&client.MicroserviceInfo{
		Name:        "clickhouse",
		Config:      "{}",
		Annotations: "{}",
		RegistryID:  5,
		Images: []client.CatalogImage{{
			ContainerImage: "dhi.io/clickhouse-server:26.7",
			ArchID:         1,
		}},
		ServiceAccount: &client.MicroserviceServiceAccountRef{
			RoleRef: client.RoleRef{
				Kind: "Role",
				Name: "microservice",
			},
		},
	}, "lima", "test-app", nil)
	require.NoError(t, err)
	require.NotNil(t, msvc.Images)
	require.Equal(t, apps.RegistryRef(5), msvc.Images.Registry)
	require.NotNil(t, msvc.ServiceAccount)
	require.Equal(t, "Role", msvc.ServiceAccount.RoleRef.Kind)
	require.Equal(t, "microservice", msvc.ServiceAccount.RoleRef.Name)

	out, err := yaml.Marshal(msvc)
	require.NoError(t, err)
	text := string(out)
	require.Contains(t, text, "registry: 5")
	require.NotContains(t, text, "registry: \"5\"")
	require.Contains(t, text, "serviceAccount:")
	require.Contains(t, text, "name: microservice")
}
