package resource

import "testing"

func TestSetTrustCA(t *testing.T) {
	ca := "dGVzdC1jYQ=="

	cp := &KubernetesControlPlane{}
	if err := SetTrustCA(cp, ca); err != nil {
		t.Fatal(err)
	}
	if cp.CA != ca {
		t.Fatalf("CA = %q", cp.CA)
	}
}

func TestSetTrustCA_unsupported(t *testing.T) {
	type fakeCP struct {
		KubernetesControlPlane
	}
	if err := SetTrustCA(&fakeCP{}, "x"); err == nil {
		t.Fatal("expected error for unsupported control plane type")
	}
}
