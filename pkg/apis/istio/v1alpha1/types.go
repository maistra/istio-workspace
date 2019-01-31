package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              *SessionSpec   `json:"spec,omitempty"`
	Status            *SessionStatus `json:"status,omitempty"`
}

type SessionSpec struct {
	Ref string `json:"ref,omitempty"`
}

type SessionStatus struct {
	State *string      `json:"state,omitempty"`
	Spec  *SessionSpec `json:"spec,omitempty"`
}
