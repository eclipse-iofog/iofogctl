package execute

import (
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/stretchr/testify/require"
)

func TestHeaderDecodeToHeaderRemoteControllerAlias(t *testing.T) {
	header := headerDecodeToHeader(&headerDecode{Kind: "RemoteController"})
	require.Equal(t, config.RemoteControllerKind, header.Kind)
}

func TestHeaderDecodeToHeaderControllerUnchanged(t *testing.T) {
	header := headerDecodeToHeader(&headerDecode{Kind: config.RemoteControllerKind})
	require.Equal(t, config.RemoteControllerKind, header.Kind)
}
