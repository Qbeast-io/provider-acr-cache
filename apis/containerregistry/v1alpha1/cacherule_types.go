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

// CacheRuleParameters are the configurable fields of a CacheRule.
type CacheRuleParameters struct {
	CacheRuleName string `json:"cacheRuleName"`
	// +kubebuilder:validation:Optional
	CredentialSetResourceId string `json:"credentialSetResourceId"`
	// +kubebuilder:validation:Optional
	CredentialSetName string `json:"credentialSetResourceName"`
	RegistryName      string `json:"registryName"`
	ResourceGroupName string `json:"resourceGroupName"`
	SourceRepository  string `json:"sourceRepository"`
	TargetRepository  string `json:"targetRepository"`
}

// CacheRuleObservation are the observable fields of a CacheRule.
type CacheRuleObservation struct {
	// The Azure API version of the resource.
	AzureApiVersion string `json:"azureApiVersion,omitempty"`
	// The creation date of the cache rule.
	CreationDate string `json:"creationDate,omitempty"`
	// The provider-assigned unique ID for this managed resource.
	Id string `json:"id,omitempty"`
	// The name of the resource.
	Name string `json:"name,omitempty"`
	// Provisioning state of the resource.
	ProvisioningState string `json:"provisioningState,omitempty"`
	// The type of the resource.
	Type               string `json:"type,omitempty"`
	CreatedByType      string `json:"createdByType,omitempty"`
	CreatedAt          string `json:"createdAt,omitempty"`
	CreatedBy          string `json:"createdBy,omitempty"`
	LastModifiedByType string `json:"lastModifiedByType,omitempty"`
	LastModifiedAt     string `json:"lastModifiedAt,omitempty"`
	LastModifiedBy     string `json:"lastModifiedBy,omitempty"`
	Ready              bool   `json:"ready,omitempty"`
}

// A CacheRuleSpec defines the desired state of a CacheRule.
type CacheRuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CacheRuleParameters `json:"forProvider"`
}

// A CacheRuleStatus represents the observed state of a CacheRule.
type CacheRuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CacheRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CacheRule is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurecontainerregistryext}
type CacheRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheRuleSpec   `json:"spec"`
	Status CacheRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CacheRuleList contains a list of CacheRule
type CacheRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRule `json:"items"`
}

// CacheRule type metadata.
var (
	CacheRuleKind             = reflect.TypeOf(CacheRule{}).Name()
	CacheRuleGroupKind        = schema.GroupKind{Group: Group, Kind: CacheRuleKind}.String()
	CacheRuleKindAPIVersion   = CacheRuleKind + "." + SchemeGroupVersion.String()
	CacheRuleGroupVersionKind = SchemeGroupVersion.WithKind(CacheRuleKind)
)

func init() {
	SchemeBuilder.Register(&CacheRule{}, &CacheRuleList{})
}
