package client

import (
	"errors"
	"testing"
	"time"
)

func TestIsAgentPlatformFailed(t *testing.T) {
	err := errors.New("agent abc platform reconcile failed: router timeout")
	if !isAgentPlatformFailed(err) {
		t.Fatal("expected platform failed detection")
	}
	if isAgentPlatformFailed(errors.New("timed out waiting for agent abc platform ready (phase=Progressing)")) {
		t.Fatal("timeout should not be treated as failed")
	}
}

func TestIsServiceProvisioningFailed(t *testing.T) {
	err := errors.New("service foo provisioning failed: hub error")
	if !isServiceProvisioningFailed(err) {
		t.Fatal("expected provisioning failed detection")
	}
	if isServiceProvisioningFailed(errors.New("timed out waiting for service foo provisioning ready (status=pending)")) {
		t.Fatal("timeout should not be treated as failed")
	}
}

func TestPlatformWaitTimeoutMatchesTests(t *testing.T) {
	const wantSeconds = 400
	if PlatformWaitTimeout != time.Duration(wantSeconds)*time.Second {
		t.Fatalf("PlatformWaitTimeout = %v, want %ds", PlatformWaitTimeout, wantSeconds)
	}
}
