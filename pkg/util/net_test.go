package util

import (
	"net"
	"testing"
)

func TestIsTCPPortOpen(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	if !IsTCPPortOpen("127.0.0.1", port) {
		t.Fatalf("expected port %d to be open", port)
	}
	if IsTCPPortOpen("127.0.0.1", 1) {
		t.Fatal("did not expect port 1 to be open")
	}
}

func TestDetectLocalHostIPv4(t *testing.T) {
	ip, err := DetectLocalHostIPv4()
	if err != nil {
		t.Skipf("DetectLocalHostIPv4 unavailable in this environment: %v", err)
	}
	parsed := net.ParseIP(ip)
	if parsed == nil || parsed.To4() == nil {
		t.Fatalf("expected IPv4 address, got %q", ip)
	}
}
