/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControlPlaneSpec defines the desired state of ControlPlane
type ControlPlaneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	User       User       `json:"user"`
	Database   Database   `json:"database,omitempty"`
	Services   Services   `json:"services,omitempty"`
	Replicas   Replicas   `json:"replicas,omitempty"`
	Images     Images     `json:"images,omitempty"`
	Ingresses  Ingresses  `json:"ingresses,omitempty"`
	Controller Controller `json:"controller,omitempty"`
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
	MessagePort  int `json:"messagePort,omitempty"`
	InteriorPort int `json:"interiorPort,omitempty"`
	EdgePort     int `json:"edgePort,omitempty"`
}

type Ingress struct {
	Address string `json:"address,omitempty"`
}

type Ingresses struct {
	Router    RouterIngress `json:"router,omitempty"`
	HTTPProxy Ingress       `json:"httpProxy,omitempty"`
	TCPProxy  Ingress       `json:"tcpProxy,omitempty"`
}

type Controller struct {
	PidBaseDir        string `json:"pidBaseDir,omitempty"`
	EcnViewerPort     int    `json:"ecnViewerPort,omitempty"`
	PortProvider      string `json:"portProvider,omitempty"`
	ECNName           string `json:"ecn,omitempty"`
	PortAllocatorHost string `json:"portAllocatorHost,omitempty"`
}

// ControlPlaneStatus defines the observed state of ControlPlane
type ControlPlaneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ControlPlane is the Schema for the controlplanes API
type ControlPlane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlPlaneSpec   `json:"spec,omitempty"`
	Status ControlPlaneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ControlPlaneList contains a list of ControlPlane
type ControlPlaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlPlane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlPlane{}, &ControlPlaneList{})
}
