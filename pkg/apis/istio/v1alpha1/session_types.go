package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file

// SessionSpec defines the desired state of Session
type SessionSpec struct {
	Route Route    `json:"route,omitempty"`
	Refs  []string `json:"ref,omitempty"`
}

// Route defines the strategy for how the traffic is routed to the Ref
type Route struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// SessionStatus defines the observed state of Session
type SessionStatus struct {
	State *string      `json:"state,omitempty"`
	Refs  []*RefStatus `json:"refs,omitempty"`
}

// RefStatus defines the observed state of the individual Ref
type RefStatus struct {
	Name      string            `json:"name,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
	Target    *RefResource      `json:"target,omitempty"`
	Resources []*RefResource    `json:"resources,omitempty"`
}

// RefResource defines the observed resources mutated/created as part of the Ref
type RefResource struct {
	Kind   *string `json:"kind,omitempty"`
	Name   *string `json:"name,omitempty"`
	Action *string `json:"action,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Session is the Schema for the sessions API
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SessionSpec   `json:"spec,omitempty"`
	Status SessionStatus `json:"status,omitempty"`
}

// HasFinalizer checks if session has a finalizer attached to it
func (s *Session) HasFinalizer(finalizer string) bool {
	for _, f := range s.Finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizer adds a finalizer to the session
func (s *Session) AddFinalizer(finalizer string) {
	s.Finalizers = append(s.Finalizers, finalizer)
}

// RemoveFinalizer removes given finalizer
func (s *Session) RemoveFinalizer(finalizer string) {
	finalizers := []string{}
	for _, f := range s.Finalizers {
		if f != finalizer {
			finalizers = append(finalizers, f)
		}
	}
	s.Finalizers = finalizers
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
