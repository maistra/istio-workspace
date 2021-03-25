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
	// How to route the given Session. A header based route using x-workspace-route with the Session name as value will be used if not provided.
	Route Route `json:"route,omitempty"`
	// Who should participate in the given session
	Refs []Ref `json:"ref,omitempty"`
}

// +k8s:openapi-gen=true
// Ref defines how to target a single Deployment or DeploymentConfig.
type Ref struct {
	// Deployment or DeploymentConfig name
	Name string `json:"name,omitempty"`
	// How this deployment should be handled, e.g. telepresence or prepared-image
	Strategy string `json:"strategy,omitempty"`
	// Additional arguments to the given strategy
	Args map[string]string `json:"args,omitempty"`
}

// +k8s:openapi-gen=true
// Route defines the strategy for how the traffic is routed to the Ref.
type Route struct {
	// The type of route to use, e.g. header
	Type string `json:"type,omitempty"`
	// Name of the key, e.g. http header
	Name string `json:"name,omitempty"`
	// The value to use for routing
	Value string `json:"value,omitempty"`
}

// +k8s:openapi-gen=true
// SessionStatus defines the observed state of Session.
type SessionStatus struct {
	State *string `json:"state,omitempty"`
	// The current configured route
	Route *Route `json:"route,omitempty"`
	// Status of the Refs in the Session
	Refs []*RefStatus `json:"refs,omitempty"`
}

// +k8s:openapi-gen=true
// Status of an individual Ref in the Session.
type RefStatus struct {
	Ref `json:",inline"`
	// A lit of the Object used as source
	Targets []*LabeledRefResource `json:"targets,omitempty"`
	// +optional
	// A list of the Resources involved in maintaining this route
	Resources []*RefResource `json:"resources,omitempty"`
}

// +k8s:openapi-gen=true
// An external Resource that has been manipulated in some way.
type RefResource struct {
	// The Resource Kind
	Kind *string `json:"kind,omitempty"`
	// The Resource Name
	Name *string `json:"name,omitempty"`
	// The Action that was performed, e.g. created, failed, manipulated
	Action *string `json:"action,omitempty"`
	// LastTransitionTime is the last time this action was applied
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
	// +optional
	// Additional properties for special Resources, e.g. hosts for Gateways
	Prop map[string]string `json:"prop,omitempty"`
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

// Session controls the creation of the specialized hidden routes.
type Session struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state
	Spec SessionSpec `json:"spec,omitempty"`
	// Status defines the current status of the State
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
