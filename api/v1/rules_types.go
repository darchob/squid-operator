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

type SquidConfig struct {
	Name string `json:"name"`
}

// RulesSpec defines the desired state of Rules
type RulesSpec struct {
	Name        string      `json:"name"`
	SquidConfig SquidConfig `json:"squidConfigRef"`
	Rules       string      `json:"rules"`
}

// RulesStatus defines the observed state of Rules
type RulesStatus struct {
	Merged bool `json:"merged,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Merged",type="boolean",JSONPath=".status.merged"
// Rules is the Schema for the rules API
type Rules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RulesSpec   `json:"spec,omitempty"`
	Status RulesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RulesList contains a list of Rules
type RulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rules{}, &RulesList{})
}
