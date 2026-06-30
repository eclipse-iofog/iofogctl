package client

import (
	"testing"

	sdkclient "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

func TestResolveRegistryID(t *testing.T) {
	tests := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"", 1, false},
		{"remote", 1, false},
		{"local", 2, false},
		{"1", 1, false},
		{"2", 2, false},
		{"10", 10, false},
		{"invalid", 0, true},
		{"0", 0, true},
	}
	for _, tt := range tests {
		got, err := ResolveRegistryID(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("ResolveRegistryID(%q) expected error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ResolveRegistryID(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Fatalf("ResolveRegistryID(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestFormatRegistryID(t *testing.T) {
	if got := FormatRegistryID(10); got != "10" {
		t.Fatalf("FormatRegistryID(10) = %q, want 10", got)
	}
}

func TestIsClientNotFoundError(t *testing.T) {
	if IsClientNotFoundError(nil) {
		t.Fatal("expected false for nil")
	}
	if !IsClientNotFoundError(sdkclient.NewNotFoundError("missing")) {
		t.Fatal("expected true for sdk NotFoundError")
	}
	if IsClientNotFoundError(sdkclient.NewInputError("bad")) {
		t.Fatal("expected false for other error")
	}
}
