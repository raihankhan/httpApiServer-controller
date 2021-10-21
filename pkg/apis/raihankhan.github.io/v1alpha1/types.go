package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ClusterID",type=string,JSONPath=`.status.klusterID`
// +kubebuilder:printcolumn:name="Progress",type=string,JSONPath=`.status.progress`

type Apiserver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiserverSpec   `json:"spec"`
	Status ApiserverStatus `json:"status,omitempty"`
}

type ApiserverSpec struct {
	DeploymentName string `json:"deployment_name"`
	ServiceName    string `json:"service_name"`
	NodePortName   string `json:"node_port_name"`
	Replicas       *int32 `json:"replicas"`
}

type ApiserverStatus struct {
	AvailableReplicas int32 `json:"available_replicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ApiserverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Apiserver `json:"items"`
}
