package configure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/internal/trust"
)

func TestControlPlaneExecutor_configureCAFromFile(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	caPath := filepath.Join(dir, "ca.pem")
	pemBytes := writeTestCAPEM(t, caPath)

	if err := config.AddNamespace("prod", "2020-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	ns, err := config.GetNamespace("prod")
	if err != nil {
		t.Fatal(err)
	}
	cp := &rsc.KubernetesControlPlane{KubeConfig: "/tmp/kube"}
	ns.SetControlPlane(cp)
	if err := config.Flush(); err != nil {
		t.Fatal(err)
	}

	exe := newControlPlaneExecutor(&Options{
		Namespace:  "prod",
		CAFile:     caPath,
		KubeConfig: "/tmp/kube-new",
	})
	if err := exe.Execute(); err != nil {
		t.Fatal(err)
	}

	stored, err := config.GetNamespace("prod")
	if err != nil {
		t.Fatal(err)
	}
	storedCP, err := stored.GetControlPlane()
	if err != nil {
		t.Fatal(err)
	}
	k8sCP, ok := storedCP.(*rsc.KubernetesControlPlane)
	if !ok {
		t.Fatal("expected kubernetes control plane")
	}
	wantCA := base64.StdEncoding.EncodeToString(pemBytes)
	if k8sCP.CA != wantCA {
		t.Fatalf("CA = %q want %q", k8sCP.CA, wantCA)
	}
	if k8sCP.KubeConfig != "/tmp/kube-new" {
		t.Fatalf("KubeConfig = %q", k8sCP.KubeConfig)
	}
	if !trust.HasCA("prod") {
		t.Fatal("expected namespace trust CA")
	}
}

func writeTestCAPEM(t *testing.T, path string) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-ca"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	return pemBytes
}
