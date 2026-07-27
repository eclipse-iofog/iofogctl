package util

import "testing"

func TestDedupeSSHTargets(t *testing.T) {
	targets := []SSHTarget{
		{User: "deploy", Host: "10.0.0.1", Port: 22, KeyFile: "/tmp/key", Label: "controller-1"},
		{User: "deploy", Host: "10.0.0.1", Port: 22, KeyFile: "/tmp/key", Label: "controller-2"},
		{User: "deploy", Host: "10.0.0.2", Port: 0, KeyFile: "/tmp/key", Label: "controller-3"},
	}

	deduped := dedupeSSHTargets(targets)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique SSH targets, got %d", len(deduped))
	}
	if deduped[1].Port != 22 {
		t.Fatalf("expected default port 22, got %d", deduped[1].Port)
	}
}
