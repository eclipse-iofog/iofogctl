package trust

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const probeDialTimeout = 10 * time.Second

// ProbeSystemTrust dials host:port with Go default root CAs and hostname verification.
func ProbeSystemTrust(ctx context.Context, host, port string) (trusted bool, err error) {
	if host == "" {
		return false, fmt.Errorf("empty TLS probe host")
	}
	if port == "" {
		port = "443"
	}
	dialer := &net.Dialer{Timeout: probeDialTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), &tls.Config{
		ServerName:         host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return false, err
	}
	_ = conn.Close()
	return true, nil
}

// IsUnknownAuthority reports whether err indicates an untrusted issuing CA.
func IsUnknownAuthority(err error) bool {
	if err == nil {
		return false
	}
	var ua x509.UnknownAuthorityError
	if errors.As(err, &ua) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unknown authority") ||
		strings.Contains(msg, "certificate signed by unknown authority")
}

// IsHostnameMismatch reports whether err indicates SAN / hostname verification failure.
func IsHostnameMismatch(err error) bool {
	if err == nil {
		return false
	}
	var hostErr x509.HostnameError
	if errors.As(err, &hostErr) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "doesn't match") ||
		strings.Contains(msg, "does not match") ||
		strings.Contains(msg, "certificate is valid for")
}
