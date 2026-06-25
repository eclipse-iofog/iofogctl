package trust

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeTrustCA_fromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ca.pem")
	writeTestCAPEM(t, path)
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	got, err := NormalizeTrustCA(path, "")
	if err != nil {
		t.Fatal(err)
	}
	want := base64.StdEncoding.EncodeToString(pemBytes)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeTrustCA_fromB64(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ca.pem")
	writeTestCAPEM(t, path)
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(pemBytes)

	got, err := NormalizeTrustCA("", encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got != encoded {
		t.Fatalf("got %q want %q", got, encoded)
	}
}

func TestNormalizeTrustCA_bothFlags(t *testing.T) {
	_, err := NormalizeTrustCA("a.pem", "dGVzdA==")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Cannot use both --ca and --ca-b64") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNormalizeTrustCA_invalidB64(t *testing.T) {
	_, err := NormalizeTrustCA("", "not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeTrustCA_invalidPEM(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("not a pem"))
	_, err := NormalizeTrustCA("", encoded)
	if err == nil {
		t.Fatal("expected error for invalid PEM content")
	}
}

func TestNormalizeTrustCA_empty(t *testing.T) {
	got, err := NormalizeTrustCA("", "")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
