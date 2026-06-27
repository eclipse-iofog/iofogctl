package deployremotecontrolplane

import (
	rsc "github.com/eclipse-iofog/iofogctl/internal/resource"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

// TranslateOptions configures RemoteControlPlane → edgelet manifest translation.
type TranslateOptions struct {
	Name            string
	Namespace       string
	ControllerImage string
	RouterImage     string
	NatsImage       string
	RegistryID      *int
}

// TranslateResult holds optional edgelet Registry YAML and required ControlPlane YAML.
type TranslateResult struct {
	Registry     []byte
	ControlPlane []byte
}

type translateOptions struct {
	name            string
	namespace       string
	controllerImage string
	routerImage     string
	natsImage       string
	registryID      *int
}

func defaultTranslateOptions(name, namespace string) translateOptions {
	return translateOptions{
		name:            name,
		namespace:       namespace,
		controllerImage: util.GetControllerImage(),
		routerImage:     util.GetRouterImage(),
		natsImage:       util.GetNatsImage(),
	}
}

func (o TranslateOptions) mergeDefaults(namespace string) translateOptions {
	base := defaultTranslateOptions(o.Name, namespace)
	if o.Name != "" {
		base.name = o.Name
	}
	if o.Namespace != "" {
		base.namespace = o.Namespace
	}
	if o.ControllerImage != "" {
		base.controllerImage = o.ControllerImage
	}
	if o.RouterImage != "" {
		base.routerImage = o.RouterImage
	}
	if o.NatsImage != "" {
		base.natsImage = o.NatsImage
	}
	base.registryID = o.RegistryID
	return base
}

// NeedsPrivateEdgeletRegistry reports whether controller.package requires an edgelet Registry manifest.
func NeedsPrivateEdgeletRegistry(cp *rsc.RemoteControlPlane) bool {
	pkg := cp.Controller.Package
	if pkg == nil {
		return false
	}
	return pkg.Registry != "" && pkg.Username != "" && pkg.Password != ""
}

// ResolveEdgeletRegistryID picks spec.controller.registry for the edgelet ControlPlane manifest.
func ResolveEdgeletRegistryID(cp *rsc.RemoteControlPlane, privateRegistryID *int) *int {
	if privateRegistryID != nil {
		return privateRegistryID
	}
	id := EdgeletRegistryOnline
	if cp != nil && cp.Airgap {
		id = EdgeletRegistryAirgap
	}
	return &id
}

// EffectiveControllerTLS returns per-controller tls override or global spec.tls.
func EffectiveControllerTLS(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController) *rsc.ControlPlaneTLS {
	if ctrl != nil && ctrl.TLS != nil {
		return ctrl.TLS
	}
	if cp != nil {
		return cp.TLS
	}
	return nil
}

// TranslateRemoteControlPlane maps global RemoteControlPlane + per-host controller to edgelet YAML.
func TranslateRemoteControlPlane(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, opts TranslateOptions) (TranslateResult, error) {
	tOpts := opts.mergeDefaults(opts.Namespace)
	var out TranslateResult

	if NeedsPrivateEdgeletRegistry(cp) {
		reg, err := translateEdgeletRegistry(cp)
		if err != nil {
			return out, err
		}
		data, err := yaml.Marshal(reg)
		if err != nil {
			return out, err
		}
		out.Registry = data
	}

	manifest := translateEdgeletControlPlane(cp, ctrl, tOpts)
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return out, err
	}
	out.ControlPlane = data
	return out, nil
}

// TranslateEdgeletControlPlaneManifest returns the edgelet ControlPlane manifest struct (test helper).
//
//nolint:revive // test helper intentionally returns package-private manifest type
func TranslateEdgeletControlPlaneManifest(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, opts TranslateOptions) edgeletControlPlaneManifest {
	return translateEdgeletControlPlane(cp, ctrl, opts.mergeDefaults(opts.Namespace))
}

// TranslateEdgeletRegistryManifest returns the edgelet Registry manifest struct (test helper).
//
//nolint:revive // test helper intentionally returns package-private manifest type
func TranslateEdgeletRegistryManifest(cp *rsc.RemoteControlPlane) (edgeletRegistryManifest, error) {
	if !NeedsPrivateEdgeletRegistry(cp) {
		return edgeletRegistryManifest{}, util.NewError("Remote Control Plane does not require a private edgelet registry")
	}
	return translateEdgeletRegistry(cp)
}

func translateEdgeletRegistry(cp *rsc.RemoteControlPlane) (edgeletRegistryManifest, error) {
	pkg := cp.Controller.Package
	return edgeletRegistryManifest{
		APIVersion: edgeletAPIVersion,
		Kind:       edgeletRegistryKind,
		Spec: edgeletRegistryManifestSpec{
			URL:      pkg.Registry,
			Username: pkg.Username,
			Password: pkg.Password,
			Email:    pkg.Email,
			Private:  true,
		},
	}, nil
}

func translateEdgeletControlPlane(cp *rsc.RemoteControlPlane, ctrl *rsc.RemoteController, opts translateOptions) edgeletControlPlaneManifest {
	sys := cp.SystemMicroservices
	spec := edgeletControlPlaneManifestSpec{
		Controller: edgeletControllerSpec{
			Image:      controllerImage(cp, opts.controllerImage),
			Registry:   opts.registryID,
			PublicURL:  resolvePublicURL(cp),
			TrustProxy: cp.Controller.TrustProxy,
		},
		Auth:     authToEdgelet(cp.Auth),
		LogLevel: cp.Controller.LogLevel,
	}
	if console := consoleToEdgelet(cp.Controller); console != nil {
		spec.Console = console
	}
	if db := databaseToEdgelet(cp.Database); db != nil {
		spec.Database = db
	}
	if ev := eventsToEdgelet(cp.Events); ev != nil {
		spec.Events = ev
	}
	if sysManifest := systemMicroservicesToEdgelet(sys, opts); sysManifest != nil {
		spec.SystemMicroservices = sysManifest
	}
	if nats := natsToEdgelet(cp.Nats); nats != nil {
		spec.Nats = nats
	}
	if tls := tlsToEdgelet(EffectiveControllerTLS(cp, ctrl)); tls != nil {
		spec.TLS = tls
	}
	if vault := vaultToEdgelet(cp.Vault); vault != nil {
		spec.Vault = vault
	}
	return edgeletControlPlaneManifest{
		APIVersion: edgeletAPIVersion,
		Kind:       edgeletControlPlaneKind,
		Metadata: edgeletControlPlaneMetadata{
			Name:      opts.name,
			Namespace: opts.namespace,
		},
		Spec: spec,
	}
}

func controllerImage(cp *rsc.RemoteControlPlane, defaultImage string) string {
	if cp.Controller.Package != nil && cp.Controller.Package.Image != "" {
		return cp.Controller.Package.Image
	}
	return defaultImage
}

func resolvePublicURL(cp *rsc.RemoteControlPlane) string {
	if cp.Controller.PublicUrl != "" {
		return cp.Controller.PublicUrl
	}
	return cp.Endpoint
}

func consoleToEdgelet(ctrl rsc.LocalControllerSpec) *edgeletConsoleSpec {
	if ctrl.ConsoleUrl == "" && ctrl.ConsolePort == 0 {
		return nil
	}
	out := &edgeletConsoleSpec{URL: ctrl.ConsoleUrl}
	if ctrl.ConsolePort != 0 {
		out.Port = &ctrl.ConsolePort
	}
	return out
}

func authToEdgelet(a rsc.Auth) edgeletAuthSpec {
	out := edgeletAuthSpec{
		Mode:                      a.Mode,
		InsecureAllowHTTP:         a.InsecureAllowHttp,
		InsecureAllowBootstrapLog: a.InsecureAllowBootstrapLog,
		IssuerURL:                 a.IssuerUrl,
		ConsoleClient:             a.ConsoleClient,
		ConsoleClientEnabled:      a.ConsoleClientEnabled,
	}
	if a.Bootstrap != nil {
		out.Bootstrap = &edgeletAuthBootstrap{
			Username: a.Bootstrap.Username,
			Password: a.Bootstrap.Password,
		}
	}
	if a.Client != nil {
		out.Client = &edgeletAuthClient{
			ID:     a.Client.ID,
			Secret: a.Client.Secret,
		}
	}
	if a.RateLimit != nil {
		out.RateLimit = &edgeletAuthRateLimit{
			Enabled:              a.RateLimit.Enabled,
			MaxRequestsPerWindow: a.RateLimit.MaxRequestsPerWindow,
			WindowMs:             a.RateLimit.WindowMs,
		}
	}
	if a.SessionStore != nil {
		out.SessionStore = &edgeletAuthSessionStore{
			Type:   a.SessionStore.Type,
			TTLMs:  a.SessionStore.TtlMs,
			Secret: a.SessionStore.Secret,
		}
	}
	if a.TokenTtl != nil {
		out.TokenTTL = &edgeletAuthTokenTTL{
			AccessTokenTTLSeconds:  a.TokenTtl.AccessTokenTtlSeconds,
			RefreshTokenTTLSeconds: a.TokenTtl.RefreshTokenTtlSeconds,
		}
	}
	if a.OidcTtl != nil {
		out.OIDCTTL = &edgeletAuthOIDCTTL{
			InteractionTTLSeconds: a.OidcTtl.InteractionTtlSeconds,
			GrantTTLSeconds:       a.OidcTtl.GrantTtlSeconds,
			SessionTTLSeconds:     a.OidcTtl.SessionTtlSeconds,
			IDTokenTTLSeconds:     a.OidcTtl.IdTokenTtlSeconds,
		}
	}
	return out
}

func databaseToEdgelet(db rsc.Database) *edgeletDatabaseSpec {
	if db.Provider == "" {
		return nil
	}
	return &edgeletDatabaseSpec{
		Provider:     db.Provider,
		User:         db.User,
		Host:         db.Host,
		Port:         db.Port,
		Password:     db.Password,
		DatabaseName: db.DatabaseName,
		SSL:          db.SSL,
		CA:           db.CA,
	}
}

func eventsToEdgelet(ev rsc.Events) *edgeletEventsSpec {
	if ev.AuditEnabled == nil && ev.RetentionDays == 0 && ev.CleanupInterval == 0 && ev.CaptureIpAddress == nil {
		return nil
	}
	return &edgeletEventsSpec{
		AuditEnabled:     ev.AuditEnabled,
		RetentionDays:    ev.RetentionDays,
		CleanupInterval:  ev.CleanupInterval,
		CaptureIPAddress: ev.CaptureIpAddress,
	}
}

func systemMicroservicesToEdgelet(sys install.RemoteSystemMicroservices, opts translateOptions) *edgeletSystemMicroservices {
	router := remoteImagesToArchMap(sys.Router, opts.routerImage)
	nats := remoteImagesToArchMap(sys.Nats, opts.natsImage)
	if len(router) == 0 && len(nats) == 0 {
		return nil
	}
	return &edgeletSystemMicroservices{
		Router: router,
		Nats:   nats,
	}
}

func remoteImagesToArchMap(images install.RemoteSystemImages, defaultImage string) map[string]string {
	out := map[string]string{}
	if images.AMD64 != "" {
		out["amd64"] = images.AMD64
	}
	if images.ARM64 != "" {
		out["arm64"] = images.ARM64
	}
	if images.RISCV64 != "" {
		out["riscv64"] = images.RISCV64
	}
	if images.ARM != "" {
		out["arm"] = images.ARM
	}
	if len(out) == 0 && defaultImage != "" {
		out = defaultArchImages(defaultImage)
	}
	return out
}

func defaultArchImages(image string) map[string]string {
	return map[string]string{
		"amd64":   image,
		"arm64":   image,
		"riscv64": image,
		"arm":     image,
	}
}

func natsToEdgelet(n *rsc.NatsEnabledConfig) *edgeletNatsSpec {
	if n == nil {
		return nil
	}
	return &edgeletNatsSpec{Enabled: n.Enabled}
}

func tlsToEdgelet(t *rsc.ControlPlaneTLS) *edgeletTLSSpec {
	if t == nil || (t.CA == "" && t.Cert == "" && t.Key == "") {
		return nil
	}
	return &edgeletTLSSpec{
		Base64: &edgeletTLSBase64{
			CA:   t.CA,
			Cert: t.Cert,
			Key:  t.Key,
		},
	}
}

func vaultToEdgelet(v *rsc.VaultSpec) *edgeletVaultSpec {
	if v == nil {
		return nil
	}
	out := &edgeletVaultSpec{
		Enabled:  v.Enabled,
		Provider: v.Provider,
		BasePath: v.BasePath,
	}
	if v.Hashicorp != nil {
		out.Hashicorp = &edgeletVaultHashicorp{
			Address: v.Hashicorp.Address,
			Token:   v.Hashicorp.Token,
			Mount:   v.Hashicorp.Mount,
		}
	}
	if v.Aws != nil {
		out.Aws = &edgeletVaultAws{
			Region:      v.Aws.Region,
			AccessKeyID: v.Aws.AccessKeyId,
			AccessKey:   v.Aws.AccessKey,
		}
	}
	if v.Azure != nil {
		out.Azure = &edgeletVaultAzure{
			URL:          v.Azure.URL,
			TenantID:     v.Azure.TenantId,
			ClientID:     v.Azure.ClientId,
			ClientSecret: v.Azure.ClientSecret,
		}
	}
	if v.Google != nil {
		out.Google = &edgeletVaultGoogle{
			ProjectID:   v.Google.ProjectId,
			Credentials: v.Google.Credentials,
		}
	}
	return out
}
