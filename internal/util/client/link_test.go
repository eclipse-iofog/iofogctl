package client

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/stretchr/testify/require"
)

func TestAgentNamesFromUUIDs(t *testing.T) {
	agents := []client.AgentInfo{
		{UUID: "aaa", Name: "lima"},
		{UUID: "bbb", Name: "edge-2"},
	}
	require.Equal(t, []string{"lima", "edge-2"}, AgentNamesFromUUIDs(agents, []string{"aaa", "bbb"}))
	require.Equal(t, []string{"lima", "missing-uuid"}, AgentNamesFromUUIDs(agents, []string{"aaa", "missing-uuid"}))
	require.Nil(t, AgentNamesFromUUIDs(agents, nil))
}

func TestPrefixRuntimeClassLink400(t *testing.T) {
	require.Nil(t, PrefixRuntimeClassLink400(nil))

	passthrough := fmt.Errorf("network")
	require.Equal(t, passthrough, PrefixRuntimeClassLink400(passthrough))

	conflict := client.NewHTTPError("bound microservice uuids: [ms-1]", 409)
	require.Equal(t, conflict, PrefixRuntimeClassLink400(conflict))
	require.Contains(t, conflict.Error(), "ms-1")

	badLink := client.NewHTTPError("RuntimeClass can only be linked to edgelet agents", 400)
	mapped := PrefixRuntimeClassLink400(badLink)
	require.Error(t, mapped)
	require.Contains(t, mapped.Error(), runtimeClassLink400Prefix)
	require.Contains(t, mapped.Error(), "RuntimeClass can only be linked to edgelet agents")
	var httpErr *client.HTTPError
	require.True(t, errors.As(mapped, &httpErr))
	require.Equal(t, 400, httpErr.Code)
}
