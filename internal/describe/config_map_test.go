package describe

import (
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestDescribeConfigMapSpec(t *testing.T) {
	spec := describeConfigMapSpec(&client.ConfigMapInfo{
		Immutable: true,
		UseVault:  false,
	})

	if spec.Immutable == nil {
		t.Fatal("Immutable is nil, want pointer to true")
	}
	if !*spec.Immutable {
		t.Fatal("Immutable = false, want true")
	}
	if spec.UseVault == nil {
		t.Fatal("UseVault is nil, want pointer to false")
	}
	if *spec.UseVault {
		t.Fatal("UseVault = true, want false")
	}
}
