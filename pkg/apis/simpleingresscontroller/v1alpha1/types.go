package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SimpleIngressRule is a specification for a SimpleIngressRule resource
type SimpleIngressRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleIngressRuleSpec   `json:"spec"`
	Status SimpleIngressRuleStatus `json:"status"`
}

// SimpleIngressRuleSpec is the spec for a Foo resource
type SimpleIngressRuleSpec struct {
	Rules []RuleSpec `json:"rules"`
}

type RuleSpec struct {
	Path    string      `json:"path"`
	Service ServiceSpec `json:"service"`
}

type ServiceSpec struct {
	Name string `json:"name"`
}

// FooStatus is the status for a Foo resource
type SimpleIngressRuleStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FooList is a list of Foo resources
type SimpleIngressRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []SimpleIngressRule `json:"items"`
}
