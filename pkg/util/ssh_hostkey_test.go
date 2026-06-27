package util

import (
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func testSSHPublicKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	return signer.PublicKey()
}

func TestKnownHostKeyFileIncludesPort(t *testing.T) {
	path22, err := knownHostKeyFile("0.0.0.0", 22)
	if err != nil {
		t.Fatalf("knownHostKeyFile: %v", err)
	}
	if !strings.HasSuffix(path22, "0.0.0.0_22.pub") {
		t.Fatalf("expected 0.0.0.0_22.pub suffix, got %q", path22)
	}

	path61954, err := knownHostKeyFile("0.0.0.0", 61954)
	if err != nil {
		t.Fatalf("knownHostKeyFile: %v", err)
	}
	if !strings.HasSuffix(path61954, "0.0.0.0_61954.pub") {
		t.Fatalf("expected 0.0.0.0_61954.pub suffix, got %q", path61954)
	}
	if path22 == path61954 {
		t.Fatal("expected different known-host paths for different ports")
	}
}

func TestVerifyHostKeyUsesCurrentPort(t *testing.T) {
	prevTerminal := sshIsTerminalFn
	sshIsTerminalFn = func(int) bool { return false }
	t.Cleanup(func() { sshIsTerminalFn = prevTerminal })

	cl := &SecureShellClient{host: "0.0.0.0", port: 22}
	cb := cl.verifyHostKey()
	cl.SetPort(61954)

	err := cb("", nil, testSSHPublicKey(t))
	if err == nil {
		t.Fatal("expected unknown host key error")
	}
	if !strings.Contains(err.Error(), "0.0.0.0:61954") {
		t.Fatalf("expected error to reference dial port 61954, got: %v", err)
	}
	if strings.Contains(err.Error(), "0.0.0.0:22") {
		t.Fatalf("expected error not to reference default port 22, got: %v", err)
	}
}

func TestVerifyHostKeyPausesSpinnerDuringPrompt(t *testing.T) {
	SpinEnable(false)
	t.Cleanup(func() { SpinEnable(true) })

	SpinStart("deploying")

	prevTerminal := sshIsTerminalFn
	sshIsTerminalFn = func(int) bool { return true }
	t.Cleanup(func() { sshIsTerminalFn = prevTerminal })

	promptCalled := false
	prevPrompt := sshHostKeyPromptFn
	sshHostKeyPromptFn = func(string, int, string) bool {
		promptCalled = true
		if isRunning {
			t.Fatal("spinner should be paused during host key prompt")
		}
		return false
	}
	t.Cleanup(func() { sshHostKeyPromptFn = prevPrompt })

	cl := &SecureShellClient{host: "example.com", port: 22}
	err := cl.verifyHostKey()("", nil, testSSHPublicKey(t))
	if err == nil {
		t.Fatal("expected verification failure")
	}
	if !promptCalled {
		t.Fatal("expected host key prompt")
	}
	if !isRunning {
		t.Fatal("spinner should resume after prompt")
	}
}
