package describe

import (
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestRoleDescribeDocumentIsControllerStyle(t *testing.T) {
	out := mustMarshal(t, roleDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleKind,
		Metadata: config.HeaderMetadata{
			Namespace: "pot-demo",
			Name:      "test-role",
		},
		Rules: []rsc.RBACRule{{
			APIGroups:     []string{""},
			Resources:     []string{"microservices", "applications"},
			Verbs:         []string{"get", "list"},
			ResourceNames: []string{"nats-test2"},
		}},
	})

	require.NotContains(t, out, "spec:")
	require.Contains(t, out, "rules:")
	require.Equal(t, 1, strings.Count(out, "name: test-role"))
	require.Equal(t, 1, strings.Count(out, "kind:"))
}

func TestRoleBindingDescribeDocumentIsControllerStyle(t *testing.T) {
	out := mustMarshal(t, roleBindingDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.RoleBindingKind,
		Metadata: config.HeaderMetadata{
			Namespace: "pot-demo",
			Name:      "bou-test-binding",
		},
		RoleRef: rsc.RoleRef{
			Kind:     "Role",
			Name:     "test-role",
			APIGroup: "datasance.com/api/v3",
		},
		Subjects: []rsc.Subject{
			{Kind: "User", Name: "emirhand89@gmail.com"},
			{Kind: "Group", Name: "test"},
		},
	})

	require.NotContains(t, out, "spec:")
	require.Contains(t, out, "roleRef:")
	require.Contains(t, out, "subjects:")
	require.NotContains(t, out, "\n  kind: RoleBinding")
}

func TestServiceAccountDescribeDocumentIsControllerStyle(t *testing.T) {
	out := mustMarshal(t, serviceAccountDocument{
		APIVersion: config.LatestAPIVersion,
		Kind:       config.ServiceAccountKind,
		Metadata: config.HeaderMetadata{
			Namespace:       "pot-demo",
			Name:            "box-debug-3",
			ApplicationName: "proxus",
		},
		RoleRef: rsc.RoleRef{Kind: "Role", Name: "microservice"},
	})

	require.NotContains(t, out, "spec:")
	require.Contains(t, out, "applicationName: proxus")
	require.Equal(t, 1, strings.Count(out, "applicationName:"))
	require.Contains(t, out, "roleRef:")
	require.NotContains(t, out, "\n  name: box-debug-3\n  applicationName:")
}

func mustMarshal(t *testing.T, doc interface{}) string {
	t.Helper()
	raw, err := yaml.Marshal(doc)
	require.NoError(t, err)
	return string(raw)
}
