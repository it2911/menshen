/*
Copyright 2019 chengchen.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type Role struct {
	Namespaces    []string `json:"namespaces,omitempty"`
	ApiGroups     []string `json:"apiGroups,omitempty"`
	Verbs         []string `json:"verbs,omitempty"`
	Resources     []string `json:"resources,omitempty"`
	NonResources  []string `json:"nonresources,omitempty"`
	ResourceNames []string `json:"resourceNames,omitempty"`
}

// RoleExtSpec defines the desired state of RoleExt
type RoleExtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Roles []Role `json:"roles,omitempty"`
}

// RoleExtStatus defines the observed state of RoleExt
type RoleExtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=roleexts,scope=Cluster

// RoleExt is the Schema for the roleexts API
type RoleExt struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleExtSpec   `json:"spec,omitempty"`
	Status RoleExtStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleExtList contains a list of RoleExt
type RoleExtList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleExt `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleExt{}, &RoleExtList{})
}
