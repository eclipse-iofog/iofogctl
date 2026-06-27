package trust

import (
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"

	"github.com/eclipse-iofog/iofogctl/internal/config"
)

func TestStoreCAAndGetCA(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	encoded := base64.StdEncoding.EncodeToString([]byte("test-ca-bytes"))

	if err := StoreCA("default", encoded); err != nil {
		t.Fatal(err)
	}
	if !HasCA("default") {
		t.Fatal("expected HasCA true")
	}
	got, err := GetCA("default")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "test-ca-bytes" {
		t.Fatalf("got %q", string(got))
	}
	mode, ok := GetCachedMode("default")
	if !ok || mode != ModeNamespace {
		t.Fatalf("mode = %q ok=%v", mode, ok)
	}

	path, _ := caPath("default")
	if filepath.Base(path) != "ca.pem" {
		t.Fatalf("unexpected ca path %s", path)
	}
}

func TestGetCA_NotFound(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	_, err := GetCA("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetCachedMode(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	if err := SetCachedMode("ns1", ModeSystem); err != nil {
		t.Fatal(err)
	}
	mode, ok := GetCachedMode("ns1")
	if !ok || mode != ModeSystem {
		t.Fatalf("mode = %q ok=%v", mode, ok)
	}
}

func TestRemoveCA(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	pem := base64.StdEncoding.EncodeToString([]byte("pem"))
	if err := StoreCA("rm", pem); err != nil {
		t.Fatal(err)
	}
	if err := RemoveCA("rm"); err != nil {
		t.Fatal(err)
	}
	if HasCA("rm") {
		t.Fatal("expected CA removed")
	}
}

func TestStoreCA_emptySkipped(t *testing.T) {
	dir := t.TempDir()
	config.Init(dir)

	if err := StoreCA("empty", "  "); err != nil {
		t.Fatal(err)
	}
	if HasCA("empty") {
		t.Fatal("empty CA should not create file")
	}
}
