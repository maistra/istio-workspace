package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// SessionSpec defines the desired state of Session
type SessionSpec struct {
	Ref string `json:"ref,omitempty"`
}

// SessionStatus defines the observed state of Session
type SessionStatus struct {
	State *string      `json:"state,omitempty"`
	Refs  []*RefStatus `json:"refs,omitempty"`
}

// RefStatus defines the observed state of the individual Ref
type RefStatus struct {
	Params    map[string]string `json:"params,omitempty"`
	Resources []*RefResource    `json:"resources,omitempty"`
}

// RefResource defines the observed resources mutated/created as part of the Ref
type RefResource struct {
	Kind *string `json:"kind,omitempty"`
	Name *string `json:"name,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Session is the Schema for the sessions API
// +k8s:openapi-gen=true
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SessionSpec   `json:"spec,omitempty"`
	Status SessionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SessionList contains a list of Session
type SessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Session `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Session{}, &SessionList{})
}
