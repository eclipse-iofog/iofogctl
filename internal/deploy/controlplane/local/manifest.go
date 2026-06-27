package deploylocalcontrolplane

const (
	edgeletAPIVersion       = "edgelet.iofog.org/v1"
	edgeletControlPlaneKind = "ControlPlane"
	edgeletRegistryKind     = "Registry"

	// EdgeletRegistryOnline is docker.io (edgelet registry ls ID 1) for online pulls.
	EdgeletRegistryOnline = 1
	// EdgeletRegistryAirgap is from_cache (edgelet registry ls ID 2) for pre-loaded images.
	EdgeletRegistryAirgap = 2
)

// edgeletControlPlaneManifest mirrors edgelet ControlPlane YAML (edgelet.iofog.org/v1).
type edgeletControlPlaneManifest struct {
	APIVersion string                          `yaml:"apiVersion"`
	Kind       string                          `yaml:"kind"`
	Metadata   edgeletControlPlaneMetadata     `yaml:"metadata"`
	Spec       edgeletControlPlaneManifestSpec `yaml:"spec"`
}

type edgeletControlPlaneMetadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

type edgeletControlPlaneManifestSpec struct {
	Controller          edgeletControllerSpec       `yaml:"controller"`
	Console             *edgeletConsoleSpec         `yaml:"console,omitempty"`
	Database            *edgeletDatabaseSpec        `yaml:"database,omitempty"`
	Auth                edgeletAuthSpec             `yaml:"auth"`
	Events              *edgeletEventsSpec          `yaml:"events,omitempty"`
	SystemMicroservices *edgeletSystemMicroservices `yaml:"systemMicroservices,omitempty"`
	Nats                *edgeletNatsSpec            `yaml:"nats,omitempty"`
	LogLevel            string                      `yaml:"logLevel,omitempty"`
	TLS                 *edgeletTLSSpec             `yaml:"tls,omitempty"`
	Vault               *edgeletVaultSpec           `yaml:"vault,omitempty"`
}

type edgeletControllerSpec struct {
	Image      string `yaml:"image"`
	Registry   *int   `yaml:"registry,omitempty"`
	Port       *int   `yaml:"port,omitempty"`
	PublicURL  string `yaml:"publicUrl,omitempty"`
	TrustProxy *bool  `yaml:"trustProxy,omitempty"`
}

type edgeletConsoleSpec struct {
	Port *int   `yaml:"port,omitempty"`
	URL  string `yaml:"url,omitempty"`
}

type edgeletDatabaseSpec struct {
	Provider     string  `yaml:"provider,omitempty"`
	User         string  `yaml:"user,omitempty"`
	Host         string  `yaml:"host,omitempty"`
	Port         int     `yaml:"port,omitempty"`
	Password     string  `yaml:"password,omitempty"`
	DatabaseName string  `yaml:"databaseName,omitempty"`
	SSL          *bool   `yaml:"ssl,omitempty"`
	CA           *string `yaml:"ca,omitempty"`
}

type edgeletAuthSpec struct {
	Mode                      string                   `yaml:"mode"`
	InsecureAllowHTTP         *bool                    `yaml:"insecureAllowHttp,omitempty"`
	InsecureAllowBootstrapLog *bool                    `yaml:"insecureAllowBootstrapLog,omitempty"`
	Bootstrap                 *edgeletAuthBootstrap    `yaml:"bootstrap,omitempty"`
	IssuerURL                 string                   `yaml:"issuerUrl,omitempty"`
	Client                    *edgeletAuthClient       `yaml:"client,omitempty"`
	ConsoleClient             string                   `yaml:"consoleClient,omitempty"`
	ConsoleClientEnabled      *bool                    `yaml:"consoleClientEnabled,omitempty"`
	RateLimit                 *edgeletAuthRateLimit    `yaml:"rateLimit,omitempty"`
	SessionStore              *edgeletAuthSessionStore `yaml:"sessionStore,omitempty"`
	TokenTTL                  *edgeletAuthTokenTTL     `yaml:"tokenTtl,omitempty"`
	OIDCTTL                   *edgeletAuthOIDCTTL      `yaml:"oidcTtl,omitempty"`
}

type edgeletAuthBootstrap struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type edgeletAuthClient struct {
	ID     string `yaml:"id,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

type edgeletAuthRateLimit struct {
	Enabled              *bool `yaml:"enabled,omitempty"`
	MaxRequestsPerWindow int   `yaml:"maxRequestsPerWindow,omitempty"`
	WindowMs             int   `yaml:"windowMs,omitempty"`
}

type edgeletAuthSessionStore struct {
	Type   string `yaml:"type,omitempty"`
	TTLMs  int    `yaml:"ttlMs,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

type edgeletAuthTokenTTL struct {
	AccessTokenTTLSeconds  int `yaml:"accessTokenTtlSeconds,omitempty"`
	RefreshTokenTTLSeconds int `yaml:"refreshTokenTtlSeconds,omitempty"`
}

type edgeletAuthOIDCTTL struct {
	InteractionTTLSeconds int `yaml:"interactionTtlSeconds,omitempty"`
	GrantTTLSeconds       int `yaml:"grantTtlSeconds,omitempty"`
	SessionTTLSeconds     int `yaml:"sessionTtlSeconds,omitempty"`
	IDTokenTTLSeconds     int `yaml:"idTokenTtlSeconds,omitempty"`
}

type edgeletEventsSpec struct {
	AuditEnabled     *bool `yaml:"auditEnabled,omitempty"`
	RetentionDays    int   `yaml:"retentionDays,omitempty"`
	CleanupInterval  int   `yaml:"cleanupInterval,omitempty"`
	CaptureIPAddress *bool `yaml:"captureIpAddress,omitempty"`
}

type edgeletSystemMicroservices struct {
	Router map[string]string `yaml:"router,omitempty"`
	Nats   map[string]string `yaml:"nats,omitempty"`
}

type edgeletNatsSpec struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

type edgeletTLSBase64 struct {
	CA   string `yaml:"ca,omitempty"`
	Cert string `yaml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

type edgeletTLSSpec struct {
	Base64 *edgeletTLSBase64 `yaml:"base64,omitempty"`
}

type edgeletVaultSpec struct {
	Enabled   *bool                  `yaml:"enabled,omitempty"`
	Provider  string                 `yaml:"provider,omitempty"`
	BasePath  string                 `yaml:"basePath,omitempty"`
	Hashicorp *edgeletVaultHashicorp `yaml:"hashicorp,omitempty"`
	Aws       *edgeletVaultAws       `yaml:"aws,omitempty"`
	Azure     *edgeletVaultAzure     `yaml:"azure,omitempty"`
	Google    *edgeletVaultGoogle    `yaml:"google,omitempty"`
}

type edgeletVaultHashicorp struct {
	Address string `yaml:"address,omitempty"`
	Token   string `yaml:"token,omitempty"`
	Mount   string `yaml:"mount,omitempty"`
}

type edgeletVaultAws struct {
	Region      string `yaml:"region,omitempty"`
	AccessKeyID string `yaml:"accessKeyId,omitempty"`
	AccessKey   string `yaml:"accessKey,omitempty"`
}

type edgeletVaultAzure struct {
	URL          string `yaml:"url,omitempty"`
	TenantID     string `yaml:"tenantId,omitempty"`
	ClientID     string `yaml:"clientId,omitempty"`
	ClientSecret string `yaml:"clientSecret,omitempty"`
}

type edgeletVaultGoogle struct {
	ProjectID   string `yaml:"projectId,omitempty"`
	Credentials string `yaml:"credentials,omitempty"`
}

type edgeletRegistryManifest struct {
	APIVersion string                      `yaml:"apiVersion"`
	Kind       string                      `yaml:"kind"`
	Spec       edgeletRegistryManifestSpec `yaml:"spec"`
}

type edgeletRegistryManifestSpec struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Private  bool   `yaml:"private"`
}
