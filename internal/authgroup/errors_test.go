package authgroup

import (
	"errors"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	t.Run("not found with name", func(t *testing.T) {
		err := MapError("secops", client.NewNotFoundError("missing"))
		require.ErrorContains(t, err, `auth group "secops" not found`)
	})

	t.Run("embedded auth required", func(t *testing.T) {
		err := MapError("", client.NewHTTPError("not implemented", 501))
		require.ErrorContains(t, err, "embedded auth mode")
	})

	t.Run("pass through unknown", func(t *testing.T) {
		base := errors.New("network down")
		require.Equal(t, base, MapError("secops", base))
	})
}
