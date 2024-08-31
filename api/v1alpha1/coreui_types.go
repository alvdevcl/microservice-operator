package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CoreUISpec defines the desired state of CoreUI
type CoreUISpec struct {
	// Add fields here
	Replicas int32  `json:"replicas"`
	Image    string `json:"image"`
}

// CoreUIStatus defines the observed state of CoreUI
type CoreUIStatus struct {
	// Add status fields here
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CoreUI is the Schema for the coreuis API
type CoreUI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CoreUISpec   `json:"spec,omitempty"`
	Status CoreUIStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CoreUIList contains a list of CoreUI
type CoreUIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CoreUI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CoreUI{}, &CoreUIList{})
}