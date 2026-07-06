package resource

import (
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/arch"
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type Microservice = apps.Microservice
type Application = apps.Application
type ApplicationTemplate = apps.ApplicationTemplate

type Container struct {
	Image       string      `yaml:"image,omitempty"`
	Credentials Credentials `yaml:"credentials,omitempty"` // Optional credentials if needed to pull images
}

type RemoteContainer struct {
	Image    string `yaml:"image,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type WasmPack struct {
	URL    string `yaml:"url,omitempty"`
	Path   string `yaml:"path,omitempty"`
	SHA256 string `yaml:"sha256,omitempty"`
}

type Package struct {
	Version   string              `yaml:"version,omitempty"`
	Container RemoteContainer     `yaml:"container,omitempty"`
	Wasm      map[string]WasmPack `yaml:"wasm,omitempty"`
}

type SSH struct {
	User    string `yaml:"user,omitempty"`
	Port    int    `yaml:"port,omitempty"`
	KeyFile string `yaml:"keyFile,omitempty"`
}

type KubeImages struct {
	PullSecret string `yaml:"pullSecret,omitempty"`
	Controller string `yaml:"controller,omitempty"`
	Operator   string `yaml:"operator,omitempty"`
	Router     string `yaml:"router,omitempty"`
	Nats       string `yaml:"nats,omitempty"`
}

type Services struct {
	Controller Service `json:"controller,omitempty" yaml:"controller,omitempty"`
	Router     Service `json:"router,omitempty" yaml:"router,omitempty"`
	Nats       Service `json:"nats,omitempty" yaml:"nats,omitempty"`
	NatsServer Service `json:"natsServer,omitempty" yaml:"natsServer,omitempty"`
}

type Service struct {
	Type                  string            `json:"type,omitempty"`
	Address               string            `json:"address,omitempty"`
	Annotations           map[string]string `json:"annotations,omitempty"`
	ExternalTrafficPolicy string            `json:"externalTrafficPolicy,omitempty" yaml:"externalTrafficPolicy,omitempty"`
}

type Replicas struct {
	Controller int32 `yaml:"controller"`
	Nats       int32 `yaml:"nats,omitempty"`
}

// Credentials credentials used to log into docker when deploying a local stack
type Credentials struct {
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Auth struct {
	Mode                      string            `yaml:"mode"`
	InsecureAllowHttp         *bool             `yaml:"insecureAllowHttp,omitempty"`
	InsecureAllowBootstrapLog *bool             `yaml:"insecureAllowBootstrapLog,omitempty"`
	Bootstrap                 *AuthBootstrap    `yaml:"bootstrap,omitempty"`
	IssuerUrl                 string            `yaml:"issuerUrl,omitempty"`
	Client                    *AuthClient       `yaml:"client,omitempty"`
	ConsoleClient             string            `yaml:"consoleClient,omitempty"`
	ConsoleClientEnabled      *bool             `yaml:"consoleClientEnabled,omitempty"`
	RateLimit                 *AuthRateLimit    `yaml:"rateLimit,omitempty"`
	SessionStore              *AuthSessionStore `yaml:"sessionStore,omitempty"`
	TokenTtl                  *AuthTokenTtl     `yaml:"tokenTtl,omitempty"`
	OidcTtl                   *AuthOidcTtl      `yaml:"oidcTtl,omitempty"`
}

type AuthBootstrap struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type AuthClient struct {
	ID     string `yaml:"id,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

type AuthRateLimit struct {
	Enabled              *bool `yaml:"enabled,omitempty"`
	MaxRequestsPerWindow int   `yaml:"maxRequestsPerWindow,omitempty"`
	WindowMs             int   `yaml:"windowMs,omitempty"`
}

type AuthSessionStore struct {
	Type   string `yaml:"type,omitempty"`
	TtlMs  int    `yaml:"ttlMs,omitempty"`
	Secret string `yaml:"secret,omitempty"`
}

type AuthTokenTtl struct {
	AccessTokenTtlSeconds  int `yaml:"accessTokenTtlSeconds,omitempty"`
	RefreshTokenTtlSeconds int `yaml:"refreshTokenTtlSeconds,omitempty"`
}

type AuthOidcTtl struct {
	InteractionTtlSeconds int `yaml:"interactionTtlSeconds,omitempty"`
	GrantTtlSeconds       int `yaml:"grantTtlSeconds,omitempty"`
	SessionTtlSeconds     int `yaml:"sessionTtlSeconds,omitempty"`
	IdTokenTtlSeconds     int `yaml:"idTokenTtlSeconds,omitempty"`
}

type Database struct {
	Provider     string  `yaml:"provider,omitempty"`
	Host         string  `yaml:"host,omitempty"`
	Port         int     `yaml:"port,omitempty"`
	User         string  `yaml:"user,omitempty"`
	Password     string  `yaml:"password,omitempty"`
	DatabaseName string  `yaml:"databaseName,omitempty"`
	SSL          *bool   `yaml:"ssl,omitempty"`
	CA           *string `yaml:"ca,omitempty"`
}

type Events struct {
	AuditEnabled     *bool `yaml:"auditEnabled,omitempty"`
	RetentionDays    int   `yaml:"retentionDays,omitempty"`
	CleanupInterval  int   `yaml:"cleanupInterval,omitempty"`
	CaptureIpAddress *bool `yaml:"captureIpAddress,omitempty"`
}

type Registry struct {
	URL          *string `yaml:"url"`
	Private      *bool   `yaml:"private"`
	Username     *string `yaml:"username"`
	Password     *string `yaml:"password"`
	Email        *string `yaml:"email"`
	RequiresCert *bool   `yaml:"requiresCert"`
	Certificate  *string `yaml:"certificate,omitempty"`
	ID           int     `yaml:"id"`
}

type Volume struct {
	Name        string   `json:"name" yaml:"name"`
	Agents      []string `json:"agents" yaml:"agents"`
	Source      string   `json:"source" yaml:"source"`
	Destination string   `json:"destination" yaml:"destination"`
	Permissions string   `json:"permissions" yaml:"permissions"`
}

type OfflineImage struct {
	Name         string            `json:"name" yaml:"name"`
	AMD64Image   string            `json:"amd64,omitempty" yaml:"amd64,omitempty"`
	ARM64Image   string            `json:"arm64,omitempty" yaml:"arm64,omitempty"`
	RISCV64Image string            `json:"riscv64,omitempty" yaml:"riscv64,omitempty"`
	ArmImage     string            `json:"arm,omitempty" yaml:"arm,omitempty"`
	Auth         *OfflineImageAuth `json:"auth,omitempty" yaml:"auth,omitempty"`
	Agents       []string          `json:"agent,omitempty" yaml:"agent,omitempty"`
}

type OfflineImageAuth struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

// AgentConfiguration contains configuration information for a deployed agent
type AgentConfiguration struct {
	Name                      string  `json:"name,omitempty" yaml:"name"`
	Location                  string  `json:"location,omitempty" yaml:"location"`
	Latitude                  float64 `json:"latitude,omitempty" yaml:"latitude"`
	Longitude                 float64 `json:"longitude,omitempty" yaml:"longitude"`
	Description               string  `json:"description,omitempty" yaml:"description"`
	Arch                      *string `json:"arch,omitempty" yaml:"arch"`
	client.AgentConfiguration `yaml:",inline"`
}

type AgentInfo struct {
	AgentConfiguration
	AgentStatus
}

type AgentStatus struct {
	LastActive            int64   `json:"lastActive" yaml:"lastActive"`
	DaemonStatus          string  `json:"daemonStatus" yaml:"daemonStatus"`
	SecurityStatus        string  `json:"securityStatus" yaml:"securityStatus"`
	SecurityViolationInfo string  `json:"securityViolationInfo" yaml:"securityViolationInfo"`
	WarningMessage        string  `json:"warningMessage" yaml:"warningMessage"`
	UptimeMs              int64   `json:"daemonOperatingDuration" yaml:"uptime"`
	MemoryUsage           float64 `json:"memoryUsage" yaml:"memoryUsage"`                     // MiB (binary) from agent status PUT
	DiskUsage             float64 `json:"diskUsage" yaml:"diskUsage"`                         // GiB (decimal) from Edgelet status PUT
	CPUUsage              float64 `json:"cpuUsage" yaml:"cpuUsage"`                           // percent
	SystemAvailableMemory float64 `json:"systemAvailableMemory" yaml:"systemAvailableMemory"` // bytes
	SystemAvailableDisk   float64 `json:"systemAvailableDisk" yaml:"systemAvailableDisk"`     // bytes
	SystemTotalCPU        float64 `json:"systemTotalCPU" yaml:"systemTotalCPU"`
	MemoryViolation       string  `json:"memoryViolation" yaml:"memoryViolation"`
	DiskViolation         string  `json:"diskViolation" yaml:"diskViolation"`
	CPUViolation          string  `json:"cpuViolation" yaml:"cpuViolation"`
	RepositoryStatus      string  `json:"repositoryStatus" yaml:"repositoryStatus"`
	LastStatusTimeMsUTC   int64   `json:"lastStatusTime" yaml:"lastStatusTime"`
	IPAddress             string  `json:"ipAddress" yaml:"ipAddress"`
	IPAddressExternal     string  `json:"ipAddressExternal" yaml:"ipAddressExternal"`
	ProcessedMessaged     int64   `json:"processedMessages" yaml:"ProcessedMessages"`
	MessageSpeed          float64 `json:"messageSpeed" yaml:"messageSpeed"`
	LastCommandTimeMsUTC  int64   `json:"lastCommandTime" yaml:"lastCommandTime"`
	Version               string  `json:"version" yaml:"version"`
	IsReadyToUpgrade      bool    `json:"isReadyToUpgrade" yaml:"isReadyToUpgrade"`
	IsReadyToRollback     bool    `json:"isReadyToRollback" yaml:"isReadyToRollback"`
	Tunnel                string  `json:"tunnel" yaml:"tunnel"`
	VolumeMounts          []VolumeMount
	GpsStatus             string                 `json:"gpsStatus" yaml:"gpsStatus"`
	AvailableRuntimes     []string               `json:"availableRuntimes" yaml:"availableRuntimes"`
	RuntimeAgentPhase     string                 `json:"runtimeAgentPhase" yaml:"runtimeAgentPhase"`
	ControlPlaneQuiesced  bool                   `json:"controlPlaneQuiesced" yaml:"controlPlaneQuiesced"`
	PlatformStatus        *client.PlatformStatus `json:"platformStatus,omitempty" yaml:"platformStatus,omitempty"`
}

// ArchStringToID maps canonical architecture names to Controller archId values.
func ArchStringToID(name string) (int64, bool) {
	id, ok := arch.NameToID[name]
	return int64(id), ok
}

// ArchIDToString maps Controller archId values to canonical architecture names.
func ArchIDToString(id int) (string, bool) {
	name, ok := arch.IDToName[id]
	return name, ok
}

// ControllerConfig is operator-aligned runtime config for the ioFog Controller (spec.controller).
type ControllerConfig struct {
	PublicUrl   string `yaml:"publicUrl,omitempty"`
	TrustProxy  *bool  `yaml:"trustProxy,omitempty"`
	ConsoleUrl  string `yaml:"consoleUrl,omitempty"`
	ConsolePort int    `yaml:"consolePort,omitempty"`
	PidBaseDir  string `yaml:"pidBaseDir,omitempty"`
	LogLevel    string `yaml:"logLevel,omitempty"`
	Https       *bool  `yaml:"https,omitempty"`
	SecretName  string `yaml:"secretName,omitempty"`
}

type RemoteControllerConfig struct {
	PidBaseDir    string           `yaml:"pidBaseDir,omitempty"`
	EcnViewerPort int              `yaml:"ecnViewerPort,omitempty"`
	EcnViewerURL  string           `yaml:"ecnViewerUrl,omitempty"`
	LogLevel      string           `yaml:"logLevel,omitempty"`
	Https         *Https           `yaml:"https,omitempty"`
	SiteCA        *SiteCertificate `yaml:"siteCA,omitempty"`  // router site CA
	LocalCA       *SiteCertificate `yaml:"localCA,omitempty"` // router local CA
}

type Https struct {
	Enabled *bool  `yaml:"enabled,omitempty"`
	CACert  string `yaml:"caCert,omitempty"`  // base64 encoded
	TLSCert string `yaml:"tlsCert,omitempty"` // base64 encoded
	TLSKey  string `yaml:"tlsKey,omitempty"`  // base64 encoded
}

type SiteCertificate struct {
	TLSCert string `yaml:"tlsCert,omitempty"` // base64 encoded
	TLSKey  string `yaml:"tlsKey,omitempty"`  // base64 encoded
}

type RouterIngress struct {
	Address      string `yaml:"address,omitempty"`
	MessagePort  int    `yaml:"messagePort,omitempty"`
	InteriorPort int    `yaml:"interiorPort,omitempty"`
	EdgePort     int    `yaml:"edgePort,omitempty"`
}

type ControllerIngress struct {
	Annotations      map[string]string `yaml:"annotations,omitempty"`
	IngressClassName string            `yaml:"ingressClassName,omitempty"`
	Host             string            `yaml:"host,omitempty"`
	SecretName       string            `yaml:"secretName,omitempty"`
}

type Ingress struct {
	Address string `yaml:"address,omitempty"`
}

// NatsIngress specifies the external address and ports for NATS hub (required when using ingress).
type NatsIngress struct {
	Address     string `yaml:"address,omitempty"`
	ServerPort  int    `yaml:"serverPort,omitempty"`
	ClusterPort int    `yaml:"clusterPort,omitempty"`
	LeafPort    int    `yaml:"leafPort,omitempty"`
	MqttPort    int    `yaml:"mqttPort,omitempty"`
	HttpPort    int    `yaml:"httpPort,omitempty"`
}

type Ingresses struct {
	Controller ControllerIngress `yaml:"controller,omitempty"`
	Router     RouterIngress     `yaml:"router,omitempty"`
	Nats       NatsIngress       `yaml:"nats,omitempty"`
}

// NatsJetStreamSpec configures JetStream storage (operator-compatible).
type NatsJetStreamSpec struct {
	StorageSize      string `yaml:"storageSize,omitempty"`
	MemoryStoreSize  string `yaml:"memoryStoreSize,omitempty"`
	StorageClassName string `yaml:"storageClassName,omitempty"`
}

// NatsSpec configures the NATS hub. When omitted, NATS is enabled with defaults.
type NatsSpec struct {
	Enabled   *bool             `yaml:"enabled,omitempty"`
	JetStream NatsJetStreamSpec `yaml:"jetStream,omitempty"`
}

// VaultSpec configures vault integration for the Controller (Kubernetes and Remote).
type VaultSpec struct {
	Enabled   *bool           `yaml:"enabled,omitempty"`
	Provider  string          `yaml:"provider,omitempty"`
	BasePath  string          `yaml:"basePath,omitempty"`
	Hashicorp *VaultHashicorp `yaml:"hashicorp,omitempty"`
	Aws       *VaultAws       `yaml:"aws,omitempty"`
	Azure     *VaultAzure     `yaml:"azure,omitempty"`
	Google    *VaultGoogle    `yaml:"google,omitempty"`
}

type VaultHashicorp struct {
	Address string `yaml:"address,omitempty"`
	Token   string `yaml:"token,omitempty"`
	Mount   string `yaml:"mount,omitempty"`
}

type VaultAws struct {
	Region      string `yaml:"region,omitempty"`
	AccessKeyId string `yaml:"accessKeyId,omitempty"`
	AccessKey   string `yaml:"accessKey,omitempty"`
}

type VaultAzure struct {
	URL          string `yaml:"url,omitempty"`
	TenantId     string `yaml:"tenantId,omitempty"`
	ClientId     string `yaml:"clientId,omitempty"`
	ClientSecret string `yaml:"clientSecret,omitempty"`
}

type VaultGoogle struct {
	ProjectId   string `yaml:"projectId,omitempty"`
	Credentials string `yaml:"credentials,omitempty"`
}

// LocalControllerSpec is operator-aligned controller config for LocalControlPlane (spec.controller).
type LocalControllerSpec struct {
	ControllerConfig `yaml:",inline"`
	Package          *ControllerPackage `yaml:"package,omitempty"`
}

// ControllerPackage holds optional controller image and private registry credentials.
type ControllerPackage struct {
	Image    string `yaml:"image,omitempty"`
	Registry string `yaml:"registry,omitempty"`
	Email    string `yaml:"email,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

// ControlPlaneTLS holds optional TLS material for non-K8s control planes.
type ControlPlaneTLS struct {
	CA   string `yaml:"ca,omitempty"`
	Cert string `yaml:"cert,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

// NatsEnabledConfig is the NATS config for remote and local control planes (enabling only; no service/ingress/jetStream).
type NatsEnabledConfig struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

// type RouterConfig struct {
// 	HA *bool `yaml:"ha,omitempty"`
// }

type Secret struct {
	Name string            `yaml:"name,omitempty"`
	Type string            `yaml:"type,omitempty"`
	Data map[string]string `yaml:"data,omitempty"`
}

type ConfigMap struct {
	Name      string            `yaml:"name,omitempty"`
	Immutable bool              `yaml:"immutable"`
	Data      map[string]string `yaml:"data,omitempty"`
}

// RBAC types (mirror SDK for describe/deploy YAML)
type RBACRule struct {
	APIGroups     []string `json:"apiGroups" yaml:"apiGroups,omitempty"`
	Resources     []string `json:"resources" yaml:"resources,omitempty"`
	Verbs         []string `json:"verbs" yaml:"verbs,omitempty"`
	ResourceNames []string `json:"resourceNames,omitempty" yaml:"resourceNames,omitempty"`
}

type RoleRef struct {
	Kind     string `json:"kind" yaml:"kind,omitempty"`
	Name     string `json:"name" yaml:"name,omitempty"`
	APIGroup string `json:"apiGroup,omitempty" yaml:"apiGroup,omitempty"`
}

type Subject struct {
	Kind     string `json:"kind" yaml:"kind,omitempty"`
	Name     string `json:"name" yaml:"name,omitempty"`
	APIGroup string `json:"apiGroup,omitempty" yaml:"apiGroup,omitempty"`
}

type Role struct {
	Name  string     `yaml:"name,omitempty"`
	Kind  string     `yaml:"kind,omitempty"`
	Rules []RBACRule `yaml:"rules,omitempty"`
}

type RoleBinding struct {
	Name     string    `yaml:"name,omitempty"`
	Kind     string    `yaml:"kind,omitempty"`
	RoleRef  RoleRef   `yaml:"roleRef,omitempty"`
	Subjects []Subject `yaml:"subjects,omitempty"`
}

// ServiceAccount is application-scoped; Controller identifies it by (applicationName, name).
type ServiceAccount struct {
	Name            string  `yaml:"name,omitempty"`
	ApplicationName string  `yaml:"applicationName,omitempty"`
	RoleRef         RoleRef `yaml:"roleRef,omitempty"`
}

type ClusterService struct {
	Name            string   `yaml:"name,omitempty"`
	Tags            []string `yaml:"tags,omitempty"`
	Type            string   `yaml:"type,omitempty"`
	Resource        string   `yaml:"resource,omitempty"`
	TargetPort      int      `yaml:"targetPort,omitempty"`
	BridgePort      int      `yaml:"bridgePort,omitempty"`
	DefaultBridge   string   `yaml:"defaultBridge,omitempty"`
	K8sType         string   `yaml:"k8sType,omitempty"`
	ServiceEndpoint string   `yaml:"serviceEndpoint,omitempty"`
	ServicePort     int      `yaml:"servicePort,omitempty"`
}

type VolumeMount struct {
	Name          string `yaml:"name,omitempty"`
	UUID          string `yaml:"uuid,omitempty"`
	ConfigMapName string `yaml:"configMapName,omitempty"`
	SecretName    string `yaml:"secretName,omitempty"`
	Version       int    `yaml:"version,omitempty"`
}

type CertificateInfo struct {
	// Name             string                 `json:"name"`
	Subject      string    `json:"subject" yaml:"subject"`
	Hosts        string    `json:"hosts" yaml:"hosts"`
	IsCA         bool      `json:"isCA" yaml:"isCA"`
	ValidFrom    time.Time `json:"validFrom" yaml:"validFrom"`
	ValidTo      time.Time `json:"validTo" yaml:"validTo"`
	SerialNumber string    `json:"serialNumber" yaml:"serialNumber"`
	CAName       *string   `json:"caName" yaml:"caName"`
	// CertificateChain []CertificateChainItem `json:"certificateChain" yaml:"certificateChain"`
	DaysRemaining int    `json:"daysRemaining" yaml:"daysRemaining"`
	IsExpired     bool   `json:"isExpired" yaml:"isExpired"`
	Certificate   string `json:"certificate" yaml:"certificate"`
	PrivateKey    string `json:"privateKey" yaml:"privateKey"`
}

type CertificateChainItem struct {
	Name    string `json:"name" yaml:"name"`
	Subject string `json:"subject" yaml:"subject"`
}

type CAInfo struct {
	// Name         string          `json:"name"`
	Subject      string    `json:"subject" yaml:"subject"`
	IsCA         bool      `json:"isCA" yaml:"isCA"`
	ValidFrom    time.Time `json:"validFrom" yaml:"validFrom"`
	ValidTo      time.Time `json:"validTo" yaml:"validTo"`
	SerialNumber string    `json:"serialNumber" yaml:"serialNumber"`
	Certificate  string    `json:"certificate" yaml:"certificate"`
	PrivateKey   string    `json:"privateKey" yaml:"privateKey"`
}

type CertificateData struct {
	Certificate string `json:"certificate" yaml:"certificate"`
	PrivateKey  string `json:"privateKey" yaml:"privateKey"`
}

// Certificate Types
type CertificateCreateRequest struct {
	Name       string              `json:"name" yaml:"name"`
	Subject    string              `json:"subject" yaml:"subject"`
	Hosts      string              `json:"hosts" yaml:"hosts"`
	Expiration int                 `json:"expiration,omitempty" yaml:"expiration,omitempty"`
	CA         CertificateCreateCA `json:"ca" yaml:"ca"`
}

type CertificateCreateResponse struct {
	Name      string    `json:"name" yaml:"name"`
	Subject   string    `json:"subject" yaml:"subject"`
	Hosts     string    `json:"hosts" yaml:"hosts"`
	ValidFrom time.Time `json:"validFrom" yaml:"validFrom"`
	ValidTo   time.Time `json:"validTo" yaml:"validTo"`
	CAName    string    `json:"caName" yaml:"caName"`
}

type CertificateCACreateResponse struct {
	Name      string    `json:"name" yaml:"name"`
	Subject   string    `json:"subject" yaml:"subject"`
	Type      string    `json:"type" yaml:"type"`
	ValidFrom time.Time `json:"validFrom" yaml:"validFrom"`
	ValidTo   time.Time `json:"validTo" yaml:"validTo"`
}

type CertificateCreateCA struct {
	Type       string `json:"type" yaml:"type"`
	SecretName string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
}

type CACreateRequest struct {
	Name       string `json:"name" yaml:"name"`
	Subject    string `json:"subject,omitempty" yaml:"subject,omitempty"`
	Expiration int    `json:"expiration,omitempty" yaml:"expiration,omitempty"`
	Type       string `json:"type" yaml:"type"`
	SecretName string `json:"secretName,omitempty" yaml:"secretName,omitempty"`
}
