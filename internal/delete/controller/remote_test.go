package deletecontroller

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoteControllerDeleteDoesNotUseLegacyUninstall(t *testing.T) {
	src, err := os.ReadFile("remote.go")
	require.NoError(t, err)
	body := string(src)
	require.NotContains(t, body, "NewController(")
	require.NotContains(t, body, "install.ControllerOptions")
	require.True(t, strings.Contains(body, "TeardownRemoteControllerHost"))
}
