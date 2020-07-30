/*
Copyright 2021.

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

// SessionSpec defines the desired state of Session.
type SessionSpec struct {
	Route Route `json:"route,omitempty"`
	Refs  []Ref `json:"ref,omitempty"`
}

// +k8s:openapi-gen=true
// Ref defines the desired state for a single reference within the Session.
type Ref struct {
	Name     string            `json:"name,omitempty"`
	Strategy string            `json:"strategy,omitempty"`
	Args     map[string]string `json:"args,omitempty"`
}

// +k8s:openapi-gen=true
// Route defines the strategy for how the traffic is routed to the Ref.
type Route struct {
	Type  string `json:"type,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// +k8s:openapi-gen=true
// SessionStatus defines the observed state of Session.
type SessionStatus struct {
	State *string      `json:"state,omitempty"`
	Route *Route       `json:"route,omitempty"`
	Refs  []*RefStatus `json:"refs,omitempty"`
}

// +k8s:openapi-gen=true
// RefStatus defines the observed state of the individual Ref.
type RefStatus struct {
	Ref       `json:",inline"`
	Targets   []*LabeledRefResource `json:"targets,omitempty"`
	Resources []*RefResource        `json:"resources,omitempty"`
}

// +k8s:openapi-gen=true
// RefResource defines the observed resources mutated/created as part of the Ref.
type RefResource struct {
	Kind   *string           `json:"kind,omitempty"`
	Name   *string           `json:"name,omitempty"`
	Action *string           `json:"action,omitempty"`
	Prop   map[string]string `json:"prop,omitempty"`
}

// +k8s:openapi-gen=true
// LabeledRefResource is a RefResource with Labels.
type LabeledRefResource struct {
	RefResource `json:",inline"`
	Labels      map[string]string `json:"labels"`
}

// +genclient
// +genclient:noStatus
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sessions,scope=Namespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Session is the Schema for the sessions API.
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SessionSpec   `json:"spec,omitempty"`
	Status SessionStatus `json:"status,omitempty"`
}

// HasFinalizer checks if session has a finalizer attached to it.
func (s *Session) HasFinalizer(finalizer string) bool {
	for _, f := range s.Finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}

// AddFinalizer adds a finalizer to the session.
func (s *Session) AddFinalizer(finalizer string) {
	s.Finalizers = append(s.Finalizers, finalizer)
}

// RemoveFinalizer removes given finalizer.
func (s *Session) RemoveFinalizer(finalizer string) {
	var finalizers []string
	for _, f := range s.Finalizers {
		if f != finalizer {
			finalizers = append(finalizers, f)
		}
	}
	s.Finalizers = finalizers
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SessionList contains a list of Session.
type SessionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Session `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Session{}, &SessionList{})
}
