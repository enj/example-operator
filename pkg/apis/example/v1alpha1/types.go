package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/api/operator/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExampleOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleOperatorSpec   `json:"status,omitempty"`
	Status ExampleOperatorStatus `json:"status,omitempty"`
}

type ExampleOperatorSpec struct {
	v1alpha1.OperatorSpec
}

type ExampleOperatorStatus struct {
	v1alpha1.OperatorStatus
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ExampleOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ExampleOperator `json:"items"`
}
