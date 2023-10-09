/*
Copyright 2023.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SSpyFamillllySpec defines the desired state of SSpyFamilllly
type SSpyFamillllySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DBHost         string `json:"dbHost,omitempty"`
	DBPort         int    `json:"dbPort,omitempty"`
	DBUser         string `json:"dbUser,omitempty"`
	DBPassword     string `json:"dbPassword,omitempty"`
	SourceDatabase string `json:"sourceDatabase,omitempty"`
	SourceTable    string `json:"sourceTable,omitempty"`
	TargetDatabase string `json:"targetDatabase,omitempty"`

	//// Markdowns contain the markdown files you want to display.
	//// The key indicates the file name and must not overlap with the keys.
	//// The value is the content in markdown format.
	////+kubebuilder:validation:Required
	////+kubebuilder:validation:MinProperties=1
	//Markdowns map[string]string `json:"markdowns,omitempty"`
	//
	//// Replicas is the number of viewers.
	//// +kubebuilder:default=1
	//// +optional
	//Replicas int32 `json:"replicas,omitempty"`
	//
	//// ViewerImage is the image name of the viewer.
	//// +optional
	//ViewerImage string `json:"viewerImage,omitempty"`
}

// SSpyFamillllyStatus defines the observed state of SSpyFamilllly
// +kubebuilder:validation:Enum=NotReady;Available;Healthy
type SSpyFamillllyStatus string

const (
	SSpyFamillllyNotReady  = SSpyFamillllyStatus("NotReady")
	SSpyFamillllyAvailable = SSpyFamillllyStatus("Available")
	SSpyFamillllyHealthy   = SSpyFamillllyStatus("Healthy")
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="REPLICAS",type="integer",JSONPath=".spec.replicas"
//+kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status"

// SSpyFamilllly is the Schema for the sspyfamillllies API
type SSpyFamilllly struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SSpyFamillllySpec   `json:"spec,omitempty"`
	Status SSpyFamillllyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SSpyFamillllyList contains a list of SSpyFamilllly
type SSpyFamillllyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SSpyFamilllly `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SSpyFamilllly{}, &SSpyFamillllyList{})
}
