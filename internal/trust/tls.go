package trust

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

// TransportConfig holds TLS settings for controller HTTP calls.
type TransportConfig struct {
	SkipVerify bool
	TLSConfig  *tls.Config
}

// TransportFromPEM builds a verifying transport from PEM bytes.
func TransportFromPEM(pem []byte) (TransportConfig, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return TransportConfig{}, fmt.Errorf("failed to parse CA certificate")
	}
	return TransportConfig{
		TLSConfig: &tls.Config{
			RootCAs:    pool,
			MinVersion: tls.VersionTLS12,
		},
	}, nil
}

// TransportFromCAFile loads a PEM file for connect-time CA override.
func TransportFromCAFile(caFile string) (TransportConfig, error) {
	pem, err := util.ReadUserFile(caFile)
	if err != nil {
		return TransportConfig{}, fmt.Errorf("read CA file: %w", err)
	}
	return TransportFromPEM(pem)
}

// ResolveConnectTransport picks TLS settings for connect using the namespace trust store.
func ResolveConnectTransport(ctx context.Context, namespace, endpoint, caFile string) (TransportConfig, error) {
	if strings.TrimSpace(caFile) != "" {
		return TransportFromCAFile(caFile)
	}
	return ResolveTransport(ctx, namespace, endpoint), nil
}

// ResolveTransport picks TLS settings for an HTTPS controller endpoint.
func ResolveTransport(ctx context.Context, namespace, endpoint string) TransportConfig {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" {
		return TransportConfig{TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12}}
	}

	if HasCA(namespace) {
		if pem, err := GetCA(namespace); err == nil {
			pool := x509.NewCertPool()
			if pool.AppendCertsFromPEM(pem) {
				return TransportConfig{
					TLSConfig: &tls.Config{
						RootCAs:    pool,
						MinVersion: tls.VersionTLS12,
					},
				}
			}
		}
	}

	if mode, ok := GetCachedMode(namespace); ok {
		switch mode {
		case ModeSystem:
			return TransportConfig{
				TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			}
		case ModeInsecure:
			return TransportConfig{
				SkipVerify: true,
				TLSConfig:  &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, // #nosec G402
			}
		}
	}

	host, port := hostPort(u)
	trusted, probeErr := ProbeSystemTrust(ctx, host, port)
	switch {
	case trusted:
		_ = SetCachedMode(namespace, ModeSystem)
		return TransportConfig{
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		}
	case IsUnknownAuthority(probeErr):
		util.PrintWarning(insecureTLSWarning(namespace,
			`Controller TLS certificate is not trusted by the system CA store for namespace "%s". Certificate verification is disabled.`,
		))
		_ = SetCachedMode(namespace, ModeInsecure)
		return TransportConfig{
			SkipVerify: true,
			TLSConfig:  &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, // #nosec G402
		}
	case IsHostnameMismatch(probeErr):
		util.PrintWarning(fmt.Sprintf(
			`Controller endpoint host "%s" does not match the TLS certificate. Check publicUrl or ingress host in your Control Plane YAML.`,
			host,
		))
		return TransportConfig{
			SkipVerify: true,
			TLSConfig:  &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, // #nosec G402
		}
	default:
		util.PrintWarning(insecureTLSWarning(namespace,
			`Controller TLS probe failed for namespace "%s". Certificate verification is disabled.`,
		))
		return TransportConfig{
			SkipVerify: true,
			TLSConfig:  &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, // #nosec G402
		}
	}
}

// TLSConfigForController resolves TLS settings for controller HTTP/WSS calls in a namespace.
func TLSConfigForController(ctx context.Context, namespace, controllerURL string) *tls.Config {
	return ResolveTransport(ctx, namespace, controllerURL).TLSConfig
}

// SDKOptions builds SDK client options with namespace-aware TLS for controller API calls.
func SDKOptions(ctx context.Context, namespace, endpoint string) (client.Options, error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return client.Options{}, err
	}
	tlsCfg := TLSConfigForController(ctx, namespace, endpoint)
	return client.Options{
		BaseURL:   baseURL,
		TLSConfig: tlsCfg,
	}, nil
}

func insecureTLSWarning(namespace, msg string) string {
	warning := fmt.Sprintf(msg, namespace)
	if !HasCA(namespace) {
		warning += ` Provide spec.ca in your Control Plane YAML, or pass connect/configure --ca <file> or --ca-b64, to enable verification.`
	}
	return warning
}

func hostPort(u *url.URL) (host, port string) {
	host = u.Hostname()
	port = u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return host, port
}
