/*
Copyright 2022.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:validation:Enum=Ready

// ConditionType represents an OriginIssuer condition value.
type ConditionType string

const (
	// ConditionReady represents that an OriginIssuer condition is in
	// a ready state and able to issue certificates.
	// If the `status` of this condition is `False`, CertificateRequest
	// controllers should prevent attempts to sign certificates.
	ConditionReady ConditionType = "Ready"
)

// +kubebuilder:validation:Enum=True;False;Unknown

const (
	// ConditionTrue represents the fact that a given condition is true.
	ConditionTrue metav1.ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false.
	ConditionFalse metav1.ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown.
	ConditionUnknown metav1.ConditionStatus = "Unknown"
)

// OCICAClusterIssuerSpec defines the desired state of OCICAClusterIssuer
type OCICAClusterIssuerSpec struct {
	// Specifies the OCID of the private CA in OCI
	OCID string `json:"ocid,omitempty"`
}

// OCICAClusterIssuerStatus defines the observed state of OCICAClusterIssuer
type OCICAClusterIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OCICAClusterIssuer is the Schema for the ocicaclusterissuers API
type OCICAClusterIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OCICAClusterIssuerSpec   `json:"spec,omitempty"`
	Status OCICAClusterIssuerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OCICAClusterIssuerList contains a list of OCICAClusterIssuer
type OCICAClusterIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OCICAClusterIssuer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OCICAClusterIssuer{}, &OCICAClusterIssuerList{})
}

func GetIssuer() (*OCICAClusterIssuer, error) {
	return &OCICAClusterIssuer{}, nil
}
