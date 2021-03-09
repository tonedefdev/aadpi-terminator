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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AzureIdentityTerminatorSpec defines the desired state of AzureIdentityTerminator
type AzureIdentityTerminatorSpec struct {
	AADRegistrationName  string `json:"aadRegistrationName"`
	AzureIdentityName    string `json:"azureIdentityName"`
	PodSelector          string `json:"podSelector"`
	ClientSecretDuration int64  `json:"clientSecretDuration"`
}

// AzureIdentityTerminatorStatus defines the observed state of AzureIdentityTerminator
type AzureIdentityTerminatorStatus struct {
	AzureIdentityBinding   string `json:"azureIdentityBinding"`
	ClientSecretExpired    bool   `json:"clientSecretExpired"`
	ClientSecretExpiration string `json:"clientSecretExpiration"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AzureIdentityTerminator is the Schema for the azureidentityterminators API
type AzureIdentityTerminator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureIdentityTerminatorSpec   `json:"spec,omitempty"`
	Status AzureIdentityTerminatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AzureIdentityTerminatorList contains a list of AzureIdentityTerminator
type AzureIdentityTerminatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureIdentityTerminator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureIdentityTerminator{}, &AzureIdentityTerminatorList{})
}
