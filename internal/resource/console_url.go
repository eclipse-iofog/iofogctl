package resource

import (
	"net/url"
	"strings"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const consoleHostPort = "80"

// ResolveConsoleURL returns the EdgeOps Console URL for a control plane.
// Primary source is spec.controller.consoleUrl. When empty, fall back to
// publicUrl, cp.Endpoint, then controllers[0].Endpoint. API-style host:port
// endpoints are mapped to the console binding on host port 80.
func ResolveConsoleURL(cp ControlPlane) (string, error) {
	if cp == nil {
		return "", util.NewInternalError("Control Plane is nil")
	}

	if consoleURL := strings.TrimSpace(controllerConsoleURL(cp)); consoleURL != "" {
		return consoleURL, nil
	}

	for _, candidate := range consoleURLFallbackCandidates(cp) {
		useHTTPS := controlPlanePrefersHTTPS(cp, candidate)
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		resolved, err := resolveConsoleFallback(candidate, useHTTPS)
		if err != nil {
			return "", err
		}
		if resolved != "" {
			return resolved, nil
		}
	}

	return "", util.NewError("Control Plane does not have a console URL")
}

// BackfillConsoleURL sets spec.controller.consoleUrl when it is empty, using ResolveConsoleURL.
func BackfillConsoleURL(cp ControlPlane) error {
	resolved, err := ResolveConsoleURL(cp)
	if err != nil {
		return err
	}
	switch c := cp.(type) {
	case *KubernetesControlPlane:
		if strings.TrimSpace(c.Controller.ConsoleUrl) == "" {
			c.Controller.ConsoleUrl = resolved
		}
	case *LocalControlPlane:
		if strings.TrimSpace(c.Controller.ConsoleUrl) == "" {
			c.Controller.ConsoleUrl = resolved
		}
	case *RemoteControlPlane:
		if strings.TrimSpace(c.Controller.ConsoleUrl) == "" {
			c.Controller.ConsoleUrl = resolved
		}
	}
	return nil
}

func controllerConsoleURL(cp ControlPlane) string {
	switch c := cp.(type) {
	case *KubernetesControlPlane:
		return c.Controller.ConsoleUrl
	case *LocalControlPlane:
		return c.Controller.ConsoleUrl
	case *RemoteControlPlane:
		return c.Controller.ConsoleUrl
	default:
		return ""
	}
}

func controllerPublicURL(cp ControlPlane) string {
	switch c := cp.(type) {
	case *KubernetesControlPlane:
		return c.Controller.PublicUrl
	case *LocalControlPlane:
		return c.Controller.PublicUrl
	case *RemoteControlPlane:
		return c.Controller.PublicUrl
	default:
		return ""
	}
}

func controlPlaneStoredEndpoint(cp ControlPlane) string {
	switch c := cp.(type) {
	case *KubernetesControlPlane:
		return c.Endpoint
	case *LocalControlPlane:
		return c.Endpoint
	case *RemoteControlPlane:
		return c.Endpoint
	default:
		return ""
	}
}

func consoleURLFallbackCandidates(cp ControlPlane) []string {
	candidates := []string{
		controllerPublicURL(cp),
		controlPlaneStoredEndpoint(cp),
	}
	controllers := cp.GetControllers()
	if len(controllers) > 0 {
		candidates = append(candidates, controllers[0].GetEndpoint())
	}
	return candidates
}

func controlPlanePrefersHTTPS(cp ControlPlane, rawURL string) bool {
	if scheme, ok := urlScheme(rawURL); ok {
		return strings.EqualFold(scheme, "https")
	}

	switch c := cp.(type) {
	case *KubernetesControlPlane:
		if c.Controller.Https != nil && *c.Controller.Https {
			return true
		}
	case *LocalControlPlane:
		if c.Controller.Https != nil && *c.Controller.Https {
			return true
		}
		if tlsEnabled(c.TLS) {
			return true
		}
	case *RemoteControlPlane:
		if c.Controller.Https != nil && *c.Controller.Https {
			return true
		}
		if tlsEnabled(c.TLS) {
			return true
		}
		for idx := range c.Controllers {
			if tlsEnabled(c.Controllers[idx].TLS) {
				return true
			}
		}
	}

	return false
}

func tlsEnabled(tls *ControlPlaneTLS) bool {
	return tls != nil && strings.TrimSpace(tls.Cert) != "" && strings.TrimSpace(tls.Key) != ""
}

func resolveConsoleFallback(raw string, useHTTPS bool) (string, error) {
	u, err := parseLooseURL(raw)
	if err != nil {
		return "", err
	}
	if u.Host == "" {
		return "", nil
	}

	if !hasNonConsolePort(u) {
		if u.Scheme == "" {
			if useHTTPS {
				u.Scheme = "https"
			} else {
				u.Scheme = "http"
			}
		}
		return u.String(), nil
	}

	host := u.Hostname()
	if useHTTPS || strings.EqualFold(u.Scheme, "https") {
		return "https://" + host + ":" + consoleHostPort, nil
	}
	return "http://" + host, nil
}

func hasNonConsolePort(u *url.URL) bool {
	port := u.Port()
	if port == "" {
		return false
	}
	return port != consoleHostPort
}

func parseLooseURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		host := raw
		if !strings.Contains(host, "://") && !strings.Contains(host, ":") {
			host = host + ":" + client.ControllerPortString
		}
		u, err = url.Parse("//" + host)
		if err != nil {
			return nil, err
		}
	}
	return u, nil
}

func urlScheme(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" {
		return "", false
	}
	return u.Scheme, true
}
