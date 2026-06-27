package trust

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTestCAPEM(t *testing.T, path string) {
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
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		t.Fatal(err)
	}
}

func TestResolveConnectTransportCAOverride(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	writeTestCAPEM(t, caPath)

	cfg, err := ResolveConnectTransport(context.Background(), "ns", "https://controller.example.com", caPath)
	if err != nil {
		t.Fatalf("ResolveConnectTransport: %v", err)
	}
	if cfg.TLSConfig == nil || cfg.TLSConfig.RootCAs == nil {
		t.Fatal("expected verifying TLS config from CA file override")
	}
	if cfg.SkipVerify {
		t.Fatal("CA override should not skip verify")
	}
}

func TestTransportFromCAFileMissing(t *testing.T) {
	_, err := TransportFromCAFile(filepath.Join(t.TempDir(), "missing.pem"))
	if err == nil {
		t.Fatal("expected error for missing CA file")
	}
}
