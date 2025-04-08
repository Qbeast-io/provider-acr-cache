/*
Copyright 2022 The Crossplane Authors.

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
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// CredentialSetParameters are the configurable fields of a CredentialSet.
type CredentialSetParameters struct {
	Name              string           `json:"name"`
	RegistryName      string           `json:"registryName"`
	Identity          Identity         `json:"identity"`
	AuthCredentials   []AuthCredential `json:"authCredentials"`
	ResourceGroupName string           `json:"resourceGroupName"`
	// +kubebuilder:validation:Optional
	LoginServer string `json:"loginServer"`
}

type Identity struct {
	Type                   string                          `json:"type"`
	PrincipalId            string                          `json:"principalId"`
	TenantId               string                          `json:"tenantId"`
	UserAssignedIdentities map[string]UserAssignedIdentity `json:"userAssignedIdentities"`
}

type UserAssignedIdentity struct {
	ClientId    string `json:"client_id"`
	PrincipalId string `json:"principal_id"`
}

type AuthCredential struct {
	Name                     string `json:"name"`
	UsernameSecretIdentifier string `json:"usernameSecretIdentifier"`
	PasswordSecretIdentifier string `json:"passwordSecretIdentifier"`
}

// CredentialSetObservation are the observable fields of a CredentialSet.
type CredentialSetObservation struct {
	// The Azure API version of the resource.
	AzureApiVersion string `json:"azureApiVersion,omitempty"`
	// The creation date of credential store resource.
	CreationDate string `json:"creationDate,omitempty"`
	// The provider-assigned unique ID for this managed resource.
	Id string `json:"id,omitempty"`
	// The name of the resource.
	Name string `json:"name,omitempty"`
	// Provisioning state of the resource.
	ProvisioningState string `json:"provisioningState,omitempty"`
	// The type of the resource.
	Type                string `json:"type,omitempty"`
	IdentityType        string `json:"identityType,omitempty"`
	IdentityTenantId    string `json:"identityTenantId,omitempty"`
	IdentityPrincipalId string `json:"identityPrincipalId,omitempty"`
	Ready               bool   `json:"ready,omitempty"`
	CreatedByType       string `json:"createdByType,omitempty"`
	CreatedAt           string `json:"createdAt,omitempty"`
	CreatedBy           string `json:"createdBy,omitempty"`
	LastModifiedByType  string `json:"lastModifiedByType,omitempty"`
	LastModifiedAt      string `json:"lastModifiedAt,omitempty"`
	LastModifiedBy      string `json:"lastModifiedBy,omitempty"`
}

type SystemDataResponse struct {
	// The timestamp of resource creation (UTC).
	CreatedAt string
	// The identity that created the resource.
	CreatedBy string
	// The type of identity that created the resource.
	CreatedByType string
	// The timestamp of resource modification (UTC).
	LastModifiedAt string
	// The identity that last modified the resource.
	LastModifiedBy string
	// The type of identity that last modified the resource.
	LastModifiedByType string
}

// A CredentialSetSpec defines the desired state of a CredentialSet.
type CredentialSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CredentialSetParameters `json:"forProvider"`
}

// A CredentialSetStatus represents the observed state of a CredentialSet.
type CredentialSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CredentialSetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CredentialSet is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurecontainerregistryext}
type CredentialSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CredentialSetSpec   `json:"spec"`
	Status CredentialSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CredentialSetList contains a list of CredentialSet
type CredentialSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CredentialSet `json:"items"`
}

// CredentialSet type metadata.
var (
	CredentialSetKind             = reflect.TypeOf(CredentialSet{}).Name()
	CredentialSetGroupKind        = schema.GroupKind{Group: Group, Kind: CredentialSetKind}.String()
	CredentialSetKindAPIVersion   = CredentialSetKind + "." + SchemeGroupVersion.String()
	CredentialSetGroupVersionKind = SchemeGroupVersion.WithKind(CredentialSetKind)
)

func init() {
	SchemeBuilder.Register(&CredentialSet{}, &CredentialSetList{})
}
