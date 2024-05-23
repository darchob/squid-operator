/*
Copyright 2024.

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

package v1

import (
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusPhase string

const (
	PhasePending  StatusPhase = "PENDING"
	PhaseDeployed StatusPhase = "DEPLOYED"
	PhaseMerged   StatusPhase = "MERGED"
	PhaseDone     StatusPhase = "DONE"
	PhaseError    StatusPhase = "ERROR"
)

// Defines the repository and tag image
type Image struct {
	// The image tag
	Tags string `json:"tag"`
	// The image repository
	Repository string `json:"repository"`
}

// ConfigsSpec defines the desired state of Configs
type ConfigsSpec struct {
	// Defines how many pods must be deployed
	Replicas *int32 `json:"replicas,omitempty"`
	// Defines the Ingress Object from `networkingv1.Ingress` object
	Ingress networkingv1.Ingress `json:"ingress,omitempty"`
	// Defines the HPA Object from `autoscalingv1.HorizontalPodAutoscaler ` object
	Hpa autoscalingv1.HorizontalPodAutoscaler `json:"hpa,omitempty"`
	// Defines the Image Object
	Image Image `json:"image,omitempty"`
	// Defines the Squid Config as String
	// squidConfig: |
	//  # Example Config
	//  http_port 3128
	//  http_access allow localnet
	//  http_access allow localhost
	//  http_access deny all
	SquidConfig string `json:"squidConfig"`
}

// ConfigsStatus defines the observed state of Configs
type ConfigsStatus struct {
	Deployment     StatusPhase `json:"deployment,omitempty"`
	ConfigMap      StatusPhase `json:"configmap,omitempty"`
	ServiceAccount StatusPhase `json:"ServiceAccount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Deployment",type="string",JSONPath=".status.deployment"
// +kubebuilder:printcolumn:name="ConfigMap",type="string",JSONPath=".status.configmap"
// +kubebuilder:printcolumn:name="ServiceAccount",type="string",JSONPath=".status.ServiceAccount"
// Configs is the Schema for the configs API
type Configs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigsSpec   `json:"spec,omitempty"`
	Status ConfigsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigsList contains a list of Configs
type ConfigsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Configs `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Configs{}, &ConfigsList{})
}
