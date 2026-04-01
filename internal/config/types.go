package config

import (
	apps "github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/apps"
)

type Kind string

const (
	AgentConfigKind            Kind = "AgentConfig"
	CatalogItemKind            Kind = "CatalogItem"
	EdgeResourceKind           Kind = "EdgeResource"
	iofogctlConfigKind         Kind = "iofogctlConfig"
	iofogctlNamespaceKind      Kind = "Namespace"
	RegistryKind               Kind = "Registry"
	VolumeKind                 Kind = "Volume"
	LocalAgentKind             Kind = "LocalAgent"
	RemoteAgentKind            Kind = "Agent"
	KubernetesControlPlaneKind Kind = "KubernetesControlPlane"
	RemoteControlPlaneKind     Kind = "ControlPlane"
	LocalControlPlaneKind      Kind = "LocalControlPlane"
	KubernetesControllerKind   Kind = "KubernetesController"
	RemoteControllerKind       Kind = "Controller"
	LocalControllerKind        Kind = "LocalController"
	MicroserviceKind           Kind = Kind(apps.MicroserviceKind)
	ApplicationKind            Kind = Kind(apps.ApplicationKind)
	ApplicationTemplateKind    Kind = Kind(apps.ApplicationTemplateKind)
	NatsAccountRuleKind        Kind = "NatsAccountRule"
	NatsUserRuleKind           Kind = "NatsUserRule"
	SecretKind                 Kind = "Secret"
	ConfigMapKind              Kind = "ConfigMap"
	ServiceKind                Kind = "Service"
	VolumeMountKind            Kind = "VolumeMount"
	OfflineImageKind           Kind = "OfflineImage"
	CertificateKind            Kind = "Certificate"
	CertificateAuthorityKind   Kind = "CertificateAuthority"
	RoleKind                   Kind = "Role"
	RoleBindingKind            Kind = "RoleBinding"
	ServiceAccountKind         Kind = "ServiceAccount"
)

// Header contains k8s yaml header
type Header struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       Kind           `yaml:"kind" json:"kind"`
	Metadata   HeaderMetadata `yaml:"metadata" json:"metadata"`
	Spec       interface{}    `yaml:"spec,omitempty" json:"spec,omitempty"`
	Data       interface{}    `yaml:"data,omitempty" json:"data,omitempty"`
	Status     interface{}    `yaml:"status,omitempty" json:"status,omitempty"`
}

// Configuration contains the unmarshalled configuration file
type configuration struct {
	DefaultNamespace string `yaml:"defaultNamespace"`
}

type iofogctlConfig struct {
	Header `yaml:",inline"`
}

type iofogctlNamespace struct {
	Header `yaml:",inline"`
}

// HeaderMetadata contains k8s metadata
type HeaderMetadata struct {
	Name            string    `yaml:"name" json:"name"`
	Namespace       string    `yaml:"namespace" json:"namespace"`
	ApplicationName string    `yaml:"applicationName,omitempty" json:"applicationName,omitempty"` // used by ServiceAccount (application-scoped)
	Tags            *[]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}
