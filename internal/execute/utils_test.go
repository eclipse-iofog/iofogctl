package execute

import (
	"bytes"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestHeaderDecodeToHeaderRemoteControllerAlias(t *testing.T) {
	header := headerDecodeToHeader(&headerDecode{Kind: "RemoteController"})
	require.Equal(t, config.RemoteControllerKind, header.Kind)
}

func TestHeaderDecodeToHeaderControllerUnchanged(t *testing.T) {
	header := headerDecodeToHeader(&headerDecode{Kind: config.RemoteControllerKind})
	require.Equal(t, config.RemoteControllerKind, header.Kind)
}

func TestHeaderDecodeAcceptsRuntimeClassRootHandler(t *testing.T) {
	raw := []byte(`apiVersion: datasance.com/v3
kind: RuntimeClass
metadata:
  name: spin
handler: spin
`)
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.SetStrict(true)
	var h headerDecode
	require.NoError(t, dec.Decode(&h))
	require.Equal(t, config.RuntimeClassKind, h.Kind)
	require.Equal(t, "spin", h.Handler)
	require.Equal(t, "spin", h.Metadata.Name)

	header := headerDecodeToHeader(&h)
	require.Equal(t, "spin", header.Handler)
	require.Equal(t, config.RuntimeClassKind, header.Kind)
}
