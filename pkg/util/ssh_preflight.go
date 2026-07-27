package util

import "fmt"

// SSHTarget identifies a remote host for SSH preflight checks.
type SSHTarget struct {
	User    string
	Host    string
	Port    int
	KeyFile string
	Label   string
}

func dedupeSSHTargets(targets []SSHTarget) []SSHTarget {
	seen := make(map[string]struct{}, len(targets))
	out := make([]SSHTarget, 0, len(targets))
	for _, target := range targets {
		port := target.Port
		if port == 0 {
			port = 22
		}
		dedupeKey := fmt.Sprintf("%s@%s:%d:%s", target.User, target.Host, port, target.KeyFile)
		if _, ok := seen[dedupeKey]; ok {
			continue
		}
		seen[dedupeKey] = struct{}{}
		normalized := target
		normalized.Port = port
		out = append(out, normalized)
	}
	return out
}

// PreflightSSHHosts serially connects to each unique SSH target so unknown host keys
// are verified before parallel deploy work begins.
func PreflightSSHHosts(targets []SSHTarget) error {
	for _, target := range dedupeSSHTargets(targets) {
		label := target.Label
		if label == "" {
			label = fmt.Sprintf("%s:%d", target.Host, target.Port)
		}

		client, err := NewSecureShellClient(target.User, target.Host, target.KeyFile)
		if err != nil {
			return fmt.Errorf("SSH preflight failed for %s: %w", label, err)
		}
		client.SetPort(target.Port)
		if err := client.Connect(); err != nil {
			return fmt.Errorf("SSH preflight failed for %s: %w", label, err)
		}
		if err := client.Disconnect(); err != nil {
			return fmt.Errorf("SSH preflight failed for %s: %w", label, err)
		}
	}
	return nil
}
