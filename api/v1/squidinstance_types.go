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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusPhase string

const (
	PhasePending  StatusPhase = "Pending"
	PhaseDeployed StatusPhase = "Deployed"
	PhaseMerged   StatusPhase = "Merged"
	PhaseDone     StatusPhase = "Done"
	PhaseError    StatusPhase = "Error"
)

// SquidInstanceSpec defines the desired state of SquidInstance
type SquidInstanceSpec struct {
	// Specify the Squid instance replicas
	Replicas *int32 `json:"replicas"`

	HpaSpec autoscalingv1.HorizontalPodAutoscalerSpec `json:"hpaSpec,omitempty"`

	StorageClassName string `json:"storageClassName"`

	// Specify the Squid container image to use
	Image *Image `json:"image"`
}

// // Defines the repository and tag image
// type JoinedConfig struct {
// 	Version string `json:"version"`
// 	Name    string `json:"name"`
// }

// Defines the repository and tag image
type Image struct {
	// The image tag
	Tag string `json:"tag"`
	// The image repository
	Repository string `json:"repository"`
}

// SquidInstanceStatus defines the observed state of SquidInstance
type SquidInstanceStatus struct {
	Health       StatusPhase               `json:"health,omitempty"`
	LoadBalancer corev1.LoadBalancerStatus `json:"loadBalancer,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Health",type="string",JSONPath=".status.health"

// SquidInstance is the Schema for the squidinstances API
type SquidInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SquidInstanceSpec   `json:"spec,omitempty"`
	Status SquidInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SquidInstanceList contains a list of SquidInstance
type SquidInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SquidInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SquidInstance{}, &SquidInstanceList{})
}
