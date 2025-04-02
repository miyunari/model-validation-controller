/*
Copyright 2025.

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

// Model defines the details of the model to validate.
type Model struct {
	Path          string `json:"path"`
	SignaturePath string `json:"signaturePath"`
}

type SigstoreConfig struct {
	CertificateIdentity   string `json:"certificateIdentity,omitempty"`
	CertificateOidcIssuer string `json:"certificateOidcIssuer,omitempty"`
}

type PkiConfig struct {
	// Path to the certificate authority for PKI.
	CertificateAuthority string `json:"certificateAuthority,omitempty"`
}

type PrivateKeyConfig struct {
	// Path to the private key.
	KeyPath string `json:"keyPath,omitempty"`
}

type ValidationConfig struct {
	SigstoreConfig   *SigstoreConfig   `json:"sigstoreConfig,omitempty"`
	PkiConfig        *PkiConfig        `json:"pkiConfig,omitempty"`
	PrivateKeyConfig *PrivateKeyConfig `json:"privateKeyConfig,omitempty"`
}

// ModelValidationSpec defines the desired state of ModelValidation
type ModelValidationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Model details.
	Model Model `json:"model"`
	// Configuration for validation methods.
	Config ValidationConfig `json:"config"`
}

// ModelValidationStatus defines the observed state of ModelValidation
type ModelValidationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ModelValidation is the Schema for the modelvalidations API
type ModelValidation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelValidationSpec   `json:"spec,omitempty"`
	Status ModelValidationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelValidationList contains a list of ModelValidation
type ModelValidationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelValidation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ModelValidation{}, &ModelValidationList{})
}
