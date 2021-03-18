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
	AppRegistration   AppRegistration  `json:"appRegistration,omitempty"`
	AzureIdentityName string           `json:"azureIdentityName"`
	NodeResourceGroup string           `json:"nodeResourceGroup"`
	PodSelector       string           `json:"podSelector"`
	ServicePrincipal  ServicePrincipal `json:"servicePrincipal,omitempty"`
}

// AzureIdentityTerminatorStatus defines the observed state of AzureIdentityTerminator
type AzureIdentityTerminatorStatus struct {
	AppRegistration      AppRegistration  `json:"appRegistration,omitempty"`
	AzureIdentityBinding string           `json:"azureIdentityBinding,omitempty"`
	RoleAssignment       RoleAssignment   `json:"roleAssignment,omitempty"`
	ServicePrincipal     ServicePrincipal `json:"servicePrincipal,omitempty"`
}

type AppRegistration struct {
	DisplayName string  `json:"displayName,omitempty"`
	ObjectID    *string `json:"objectID,omitempty"`
}

type RoleAssignment struct {
	Name     *string `json:"name,omitempty"`
	ObjectID *string `json:"objectID,omitempty"`
}

type ServicePrincipal struct {
	ClientSecretDuration   string       `json:"clientSecretDuration,omitempty"`
	ClientSecretExpiration *metav1.Time `json:"clientSecretExpiration,omitempty"`
	ObjectID               *string      `json:"objectID,omitempty"`
	Tags                   []string     `json:"tags,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="azidt"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AADApplication",type="string",JSONPath=".spec.appRegistration.displayName",description="The name of the Azure AD Application registered"
// +kubebuilder:printcolumn:name="ClientSecretDuration",type="string",JSONPath=".spec.servicePrincipal.clientSecretDuration",description="The life time of the ClientSecret"
// +kubebuilder:printcolumn:name="ClientSecretExp",type="string",JSONPath=".status.servicePrincipal.clientSecretExpiration",description="The time the ClientSecret will expire"
// +kubebuilder:printcolumn:name="PodSelector",type="string",JSONPath=".spec.podSelector",description="The selector that will bind pods to the AzureIdentityBinding"
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
