package install

import (
	cpv3 "github.com/eclipse-iofog/iofog-operator/v3/apis/controlplanes/v3"
)

type IofogUser struct {
	Name         string
	Surname      string
	Email        string
	Password     string
	AccessToken  string
	RefreshToken string
}

type Auth = cpv3.Auth

type Database struct {
	Provider     string
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSL          *bool
	CA           *string
}

type Events struct {
	AuditEnabled     *bool
	RetentionDays    int
	CleanupInterval  int
	CaptureIpAddress *bool
}

// VaultConfig is used by the remote controller install to pass VAULT_* env vars (operator-compatible shape).
type VaultConfig struct {
	Enabled   *bool
	Provider  string
	BasePath  string
	Hashicorp *VaultHashicorpConfig
	Aws       *VaultAwsConfig
	Azure     *VaultAzureConfig
	Google    *VaultGoogleConfig
}

type VaultHashicorpConfig struct {
	Address string
	Token   string
	Mount   string
}

type VaultAwsConfig struct {
	Region      string
	AccessKeyId string
	AccessKey   string
}

type VaultAzureConfig struct {
	URL          string
	TenantId     string
	ClientId     string
	ClientSecret string
}

type VaultGoogleConfig struct {
	ProjectId   string
	Credentials string
}

type RemoteSystemImages struct {
	ARM     string `yaml:"arm,omitempty"`
	AMD64   string `yaml:"amd64,omitempty"`
	ARM64   string `yaml:"arm64,omitempty"`
	RISCV64 string `yaml:"riscv64,omitempty"`
}

type RemoteSystemMicroservices struct {
	Router RemoteSystemImages `yaml:"router,omitempty"`
	Nats   RemoteSystemImages `yaml:"nats,omitempty"`
}

type SiteCertificate struct {
	TLSCert string
	TLSKey  string
}

type Pod struct {
	Name   string
	Status string
}

type K8SControllerConfig struct {
	Replicas     int32
	ReplicasNats int32
	Database     Database
	Auth         Auth
	Events       Events
	Controller   cpv3.Controller
	Nats         *cpv3.Nats
	Vault        *cpv3.Vault
}
