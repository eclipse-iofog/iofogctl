package iofog

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControlPlaneSpec defines the desired state of ControlPlane
// +k8s:openapi-gen=true
type ControlPlaneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	User      User      `json:"user"`
	Database  Database  `json:"database,omitempty"`
	Services  Services  `json:"services,omitempty"`
	Replicas  Replicas  `json:"replicas,omitempty"`
	Images    Images    `json:"images,omitempty"`
	Ingresses Ingresses `json:"ingresses,omitempty"`
}

type Replicas struct {
	Controller int32 `json:"controller,omitempty"`
}

type Services struct {
	Controller Service `json:"controller,omitempty"`
	Router     Service `json:"router,omitempty"`
	Proxy      Service `json:"proxy,omitempty"`
}

type Service struct {
	Type    string `json:"type,omitempty"`
	Address string `json:"address,omitempty"`
}

type Images struct {
	PullSecret  string `json:"pullSecret,omitempty"`
	Kubelet     string `json:"kubelet,omitempty"`
	Controller  string `json:"controller,omitempty"`
	Router      string `json:"router,omitempty"`
	PortManager string `json:"portManager,omitempty"`
	Proxy       string `json:"proxy,omitempty"`
}

type Database struct {
	Provider     string `json:"provider"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
}

type User struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RouterIngress struct {
	Ingress
	HttpPort     int `json:"httpPort,omitempty"`
	MessagePort  int `json:"messagePort,omitempty"`
	InteriorPort int `json:"interiorPort,omitempty"`
	EdgePort     int `json:"edgePort,omitempty"`
}

type TcpIngress struct {
	Ingress
	TcpAllocatorHost string `json:"tcpAllocatorHost,omitempty"`
	TcpAllocatorPort int    `json:"tcpAllocatorPort,omitempty"`
	EcnId            int    `json:"ecnId,omitempty"`
}

type Ingress struct {
	Address string `json:"address,omitempty"`
}

type Ingresses struct {
	Router    RouterIngress `json:"router,omitempty"`
	HttpProxy Ingress       `json:"httpProxy,omitempty"`
	TcpProxy  TcpIngress    `json:"tcpProxy,omitempty"`
}

// ControlPlaneStatus defines the observed state of ControlPlane
// +k8s:openapi-gen=true
type ControlPlaneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlane is the Schema for the control plane API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ControlPlaneList contains a list of ControlPlane
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
