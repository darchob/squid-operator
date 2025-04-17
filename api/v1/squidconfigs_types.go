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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SquidConfigsSpec defines the desired state of SquidConfigs
type SquidConfigsSpec struct {
	Rules string `json:"rules"`
}

// SquidInstanceRef defines the Squid instance to apply the configuration to
type SquidInstanceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// SquidConfigsStatus defines the observed state of SquidConfigs
type SquidConfigsStatus struct {
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Merged",type="boolean",JSONPath=".status.merged"
// +kubebuilder:printcolumn:name="Applied",type="boolean",JSONPath=".status.apply"

// SquidConfigs is the Schema for the squidconfigs API
type SquidConfigs struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SquidConfigsSpec   `json:"spec,omitempty"`
	Status SquidConfigsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SquidConfigsList contains a list of SquidConfigs
type SquidConfigsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SquidConfigs `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SquidConfigs{}, &SquidConfigsList{})
}
