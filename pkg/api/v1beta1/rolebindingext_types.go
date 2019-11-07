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

type Subject struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
}

// RoleBindingExtSpec defines the desired state of RoleBindingExt
type RoleBindingExtSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Subjects  []Subject `json:"subjects,omitempty"`
	RoleNames []string  `json:"roleNames,omitempty"`
	Message   string    `json:"message,omitempty"`
	Type      string    `json:"type,omitempty"`    // allow or deny
	Crontab   string    `json:"crontab,omitempty"` //
}

// RoleBindingExtStatus defines the observed state of RoleBindingExt
type RoleBindingExtStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// RoleBindingExt is the Schema for the rolebindingexts API
type RoleBindingExt struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleBindingExtSpec   `json:"spec,omitempty"`
	Status RoleBindingExtStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleBindingExtList contains a list of RoleBindingExt
type RoleBindingExtList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleBindingExt `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RoleBindingExt{}, &RoleBindingExtList{})
}
