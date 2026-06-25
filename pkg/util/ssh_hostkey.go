package util

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	sshHostKeyPromptFn = defaultSSHHostKeyPrompt
	sshIsTerminalFn    = term.IsTerminal
	sshStdin           = os.Stdin
	sshStdout          = os.Stdout
)

func knownHostsRoot() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".iofog", "v3", "known_hosts"), nil
}

func knownHostKeyFile(host string, port int) (string, error) {
	root, err := knownHostsRoot()
	if err != nil {
		return "", err
	}
	safeHost := strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(host)
	name := fmt.Sprintf("%s_%d.pub", safeHost, port)
	return filepath.Join(root, name), nil
}

func loadKnownHostKey(path string) (ssh.PublicKey, error) {
	root, err := knownHostsRoot()
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, err
	}
	data, err := ReadFileUnderRoot(root, rel)
	if err != nil {
		return nil, err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, err
	}
	return ssh.ParsePublicKey(raw)
}

func saveKnownHostKey(path string, key ssh.PublicKey) error {
	if err := os.MkdirAll(filepath.Dir(path), DirPerm); err != nil {
		return err
	}
	encoded := base64.StdEncoding.EncodeToString(key.Marshal())
	return os.WriteFile(path, []byte(encoded), FilePerm)
}

func (cl *SecureShellClient) verifyHostKey() ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		path, err := knownHostKeyFile(cl.host, cl.port)
		if err != nil {
			return err
		}
		stored, err := loadKnownHostKey(path)
		if err == nil {
			if string(stored.Marshal()) != string(key.Marshal()) {
				return fmt.Errorf("host key for %s:%d changed; remove %s to re-trust", cl.host, cl.port, path)
			}
			return nil
		}
		if !os.IsNotExist(err) {
			return err
		}
		if !sshIsTerminalFn(int(sshStdin.Fd())) {
			return fmt.Errorf("unknown SSH host key for %s:%d; run interactively to verify fingerprint", cl.host, cl.port)
		}
		SpinHandlePrompt()
		defer SpinHandlePromptComplete()

		fingerprint := ssh.FingerprintSHA256(key)
		if !sshHostKeyPromptFn(cl.host, cl.port, fingerprint) {
			return fmt.Errorf("host key verification failed for %s:%d", cl.host, cl.port)
		}
		return saveKnownHostKey(path, key)
	}
}

func defaultSSHHostKeyPrompt(host string, port int, fingerprint string) bool {
	fmt.Fprintf(sshStdout, "The authenticity of host '%s:%d' can't be established.\n", host, port)
	fmt.Fprintf(sshStdout, "ED25519/EC/RSA key fingerprint is SHA256:%s.\n", fingerprint)
	fmt.Fprint(sshStdout, "Are you sure you want to continue connecting (yes/no)? ")
	reader := bufio.NewReader(sshStdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "yes" || answer == "y"
}
