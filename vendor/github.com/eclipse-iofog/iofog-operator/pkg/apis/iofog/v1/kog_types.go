package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KogSpec defines the desired state of Kog
// +k8s:openapi-gen=true
type KogSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ControlPlane ControlPlane `json:"controlPlane"`
	Connectors   Connectors   `json:"connectors,omitempty"`
}

type Connectors struct {
	Image     string      `json:"image"`
	Instances []Connector `json:"instances"`
}

type Connector struct {
	Name string `json:"name"`
}

type ControlPlane struct {
	IofogUser              IofogUser `json:"iofogUser"`
	ControllerReplicaCount int32     `json:"controllerReplicaCount"`
	Database               Database  `json:"database,omitempty"`
	ControllerImage        string    `json:"controllerImage"`
	ImagePullSecret        string    `json:"imagePullSecret,omitempty"`
	KubeletImage           string    `json:"kubeletImage"`
	ServiceType            string    `json:"serviceType"`
	LoadBalancerIP         string    `json:"loadBalancerIp,omitempty"`
}

type Database struct {
	Provider     string `json:"provider"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DatabaseName string `json:"databaseName"`
}

type IofogUser struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// KogStatus defines the observed state of Kog
// +k8s:openapi-gen=true
type KogStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	ControllerPods []string `json:"controllerPods"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Kog is the Schema for the kogs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Kog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KogSpec   `json:"spec,omitempty"`
	Status KogStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KogList contains a list of Kog
type KogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kog `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kog{}, &KogList{})
}
