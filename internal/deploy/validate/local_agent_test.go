package validate

import (
	"net"
	"testing"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestLocalAgentPortAvailable_SkipsForSystemAgent(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:54321")
	if err != nil {
		t.Skipf("cannot bind 127.0.0.1:54321 for test: %v", err)
	}
	defer ln.Close()

	require.NoError(t, LocalAgentPortAvailable(true))
}

func TestLocalAgentPortAvailable_RejectsWhenPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:54321")
	if err != nil {
		t.Skipf("cannot bind 127.0.0.1:54321 for test: %v", err)
	}
	defer ln.Close()

	err = LocalAgentPortAvailable(false)
	var conflictErr *util.ConflictError
	require.ErrorAs(t, err, &conflictErr)
	require.Contains(t, err.Error(), "54321")
}

func TestLocalAgentPortAvailable_AllowsWhenPortFree(t *testing.T) {
	if util.IsTCPPortOpen("127.0.0.1", LocalAgentConflictPort) {
		t.Skip("port 54321 already in use on this host")
	}
	require.NoError(t, LocalAgentPortAvailable(false))
}
