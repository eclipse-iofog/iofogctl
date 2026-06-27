package resource

import (
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
)

// AuthToCPV3 converts CLI resource auth to operator ControlPlane auth (3D translator helper).
func AuthToCPV3(a Auth) cpv3.Auth {
	out := cpv3.Auth{
		Mode:                      cpv3.AuthMode(a.Mode),
		InsecureAllowHttp:         a.InsecureAllowHttp,
		InsecureAllowBootstrapLog: a.InsecureAllowBootstrapLog,
		IssuerUrl:                 a.IssuerUrl,
		ConsoleClient:             a.ConsoleClient,
		ConsoleClientEnabled:      a.ConsoleClientEnabled,
	}
	if a.Bootstrap != nil {
		out.Bootstrap = &cpv3.AuthBootstrap{
			Username: a.Bootstrap.Username,
			Password: a.Bootstrap.Password,
		}
	}
	if a.Client != nil {
		out.Client = &cpv3.AuthClient{
			ID:     a.Client.ID,
			Secret: a.Client.Secret,
		}
	}
	if a.RateLimit != nil {
		out.RateLimit = &cpv3.AuthRateLimit{
			Enabled:              a.RateLimit.Enabled,
			MaxRequestsPerWindow: a.RateLimit.MaxRequestsPerWindow,
			WindowMs:             a.RateLimit.WindowMs,
		}
	}
	if a.SessionStore != nil {
		out.SessionStore = &cpv3.AuthSessionStore{
			Type:   a.SessionStore.Type,
			TtlMs:  a.SessionStore.TtlMs,
			Secret: a.SessionStore.Secret,
		}
	}
	if a.TokenTtl != nil {
		out.TokenTtl = &cpv3.AuthTokenTtl{
			AccessTokenTtlSeconds:  a.TokenTtl.AccessTokenTtlSeconds,
			RefreshTokenTtlSeconds: a.TokenTtl.RefreshTokenTtlSeconds,
		}
	}
	if a.OidcTtl != nil {
		out.OidcTtl = &cpv3.AuthOidcTtl{
			InteractionTtlSeconds: a.OidcTtl.InteractionTtlSeconds,
			GrantTtlSeconds:       a.OidcTtl.GrantTtlSeconds,
			SessionTtlSeconds:     a.OidcTtl.SessionTtlSeconds,
			IdTokenTtlSeconds:     a.OidcTtl.IdTokenTtlSeconds,
		}
	}
	return out
}

// ControllerConfigToCPV3 converts spec.controller to operator Controller (3D translator helper).
func ControllerConfigToCPV3(c ControllerConfig) cpv3.Controller {
	https := c.Https
	if https == nil {
		defaultHTTPS := false
		https = &defaultHTTPS
	}
	return cpv3.Controller{
		PublicUrl:   c.PublicUrl,
		TrustProxy:  c.TrustProxy,
		ConsoleUrl:  c.ConsoleUrl,
		ConsolePort: c.ConsolePort,
		PidBaseDir:  c.PidBaseDir,
		Https:       https,
		SecretName:  c.SecretName,
		LogLevel:    c.LogLevel,
	}
}
