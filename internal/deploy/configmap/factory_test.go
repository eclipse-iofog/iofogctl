package deployconfigmap

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestImmutableForCreate(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		if got := immutableForCreate(rsc.ConfigMap{}); got != nil {
			t.Fatalf("immutableForCreate() = %v, want nil", got)
		}
	})

	t.Run("explicit false", func(t *testing.T) {
		got := immutableForCreate(rsc.ConfigMap{Immutable: boolPtr(false)})
		if got == nil || *got {
			t.Fatalf("immutableForCreate() = %v, want false", got)
		}
	})
}

func TestImmutableForUpdate(t *testing.T) {
	existing := &client.ConfigMapInfo{Immutable: true}

	t.Run("omitted uses existing", func(t *testing.T) {
		got := immutableForUpdate(rsc.ConfigMap{}, existing)
		if got == nil || !*got {
			t.Fatalf("immutableForUpdate() = %v, want true", got)
		}
	})

	t.Run("explicit false", func(t *testing.T) {
		got := immutableForUpdate(rsc.ConfigMap{Immutable: boolPtr(false)}, existing)
		if got == nil || *got {
			t.Fatalf("immutableForUpdate() = %v, want false", got)
		}
	})
}

func TestUseVaultForCreate(t *testing.T) {
	t.Run("omitted", func(t *testing.T) {
		if got := useVaultForCreate(rsc.ConfigMap{}); got != nil {
			t.Fatalf("useVaultForCreate() = %v, want nil", got)
		}
	})

	t.Run("explicit false", func(t *testing.T) {
		got := useVaultForCreate(rsc.ConfigMap{UseVault: boolPtr(false)})
		if got == nil || *got {
			t.Fatalf("useVaultForCreate() = %v, want false", got)
		}
	})

	t.Run("explicit true", func(t *testing.T) {
		got := useVaultForCreate(rsc.ConfigMap{UseVault: boolPtr(true)})
		if got == nil || !*got {
			t.Fatalf("useVaultForCreate() = %v, want true", got)
		}
	})
}

func TestUseVaultForUpdate(t *testing.T) {
	existing := &client.ConfigMapInfo{UseVault: true}

	t.Run("omitted", func(t *testing.T) {
		if got := useVaultForUpdate(rsc.ConfigMap{}, existing); got != nil {
			t.Fatalf("useVaultForUpdate() = %v, want nil", got)
		}
	})

	t.Run("unchanged true", func(t *testing.T) {
		if got := useVaultForUpdate(rsc.ConfigMap{UseVault: boolPtr(true)}, existing); got != nil {
			t.Fatalf("useVaultForUpdate() = %v, want nil", got)
		}
	})

	t.Run("changed to false", func(t *testing.T) {
		got := useVaultForUpdate(rsc.ConfigMap{UseVault: boolPtr(false)}, existing)
		if got == nil || *got {
			t.Fatalf("useVaultForUpdate() = %v, want false", got)
		}
	})

	t.Run("changed to true", func(t *testing.T) {
		existingFalse := &client.ConfigMapInfo{UseVault: false}
		got := useVaultForUpdate(rsc.ConfigMap{UseVault: boolPtr(true)}, existingFalse)
		if got == nil || !*got {
			t.Fatalf("useVaultForUpdate() = %v, want true", got)
		}
	})
}
