package get

import (
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAllRoutinesIncludeModelsAndKnowledge(t *testing.T) {
	t.Parallel()

	names := make([]string, 0, len(routines))
	for _, routine := range routines {
		names = append(names, runtime.FuncForPC(reflect.ValueOf(routine).Pointer()).Name())
	}
	joined := strings.Join(names, "\n")
	require.Contains(t, joined, "getModelTable")
	require.Contains(t, joined, "getKnowledgeTable")
	require.NotContains(t, joined, "getRuntimeClass")
	require.NotContains(t, joined, "getMicroserviceTemplate")
}
