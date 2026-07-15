package nodeversion

import (
	"errors"
	"strings"
	"testing"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestMapErrorUpgradeNotReady(t *testing.T) {
	err := MapError("upgrade", client.NewHTTPError("INVALID_VERSION_COMMAND_UPGRADE", 400))
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "agent is not ready to upgrade" {
		t.Fatalf("got %q", got)
	}
}

func TestMapErrorRollbackNotReady(t *testing.T) {
	err := MapError("rollback", client.NewHTTPError("INVALID_VERSION_COMMAND_ROLLBACK", 400))
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "agent is not ready to rollback" {
		t.Fatalf("got %q", got)
	}
}

func TestMapErrorInvalidSemver(t *testing.T) {
	err := MapError("upgrade", client.NewHTTPError("ValidationError", 400))
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); !strings.Contains(got, "invalid semver for upgrade") {
		t.Fatalf("got %q", got)
	}
}

func TestMapErrorPassthrough(t *testing.T) {
	orig := errors.New("network down")
	if got := MapError("upgrade", orig); !errors.Is(got, orig) {
		t.Fatalf("expected passthrough, got %v", got)
	}
}

func TestSemverPtr(t *testing.T) {
	if SemverPtr("") != nil {
		t.Fatal("expected nil for empty semver")
	}
	semver := "3.2.0"
	ptr := SemverPtr(semver)
	if ptr == nil || *ptr != semver {
		t.Fatalf("got %v", ptr)
	}
}
