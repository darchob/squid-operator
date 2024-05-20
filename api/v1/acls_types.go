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

// AclsSpec defines the desired state of Acls
type AclsSpec struct {
	// Defines the Squid ACLs Config as String
	// squidConfig: |
	//  # Example Config
	//  acl H2G2 10.0.0.0/8
	//  http_access allow H2G2
	//  ...
	ACLs string `json:"acls"`
}

// AclsStatus defines the observed state of Acls
type AclsStatus struct {
	// Ensure ACLs is merged to the global configmap
	Merged bool `json:"merged,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Acls is the Schema for the acls API
type Acls struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AclsSpec   `json:"spec,omitempty"`
	Status AclsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AclsList contains a list of Acls
type AclsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Acls `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Acls{}, &AclsList{})
}
