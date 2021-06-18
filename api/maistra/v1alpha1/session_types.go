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
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SessionState string

const (
	// StateProcessing indicates that particular action related to the session is ongoing.
	StateProcessing SessionState = "Processing"
	// StateSuccess indicates that particular action related to the session has finished successfully.
	StateSuccess SessionState = "Success"
	// StateFailed indicates that particular action related to the session has failed.
	StateFailed SessionState = "Failed"
	// StatusFailed indicates that overall condition failed
	StatusFailed string = "false"
)

// SessionSpec defines the desired state of Session.
type SessionSpec struct {
	// How to route the given Session. A header based route using x-workspace-route with the Session name as value will be used if not provided.
	Route Route `json:"route,omitempty"`
	// Who should participate in the given session
	Refs []Ref `json:"ref,omitempty"`
}

// Ref defines how to target a single Deployment or DeploymentConfig.
// +k8s:openapi-gen=true
type Ref struct {
	// Deployment or DeploymentConfig name, could optionally contain [Kind/]Name to be specific
	Name string `json:"name,omitempty"`
	// How this deployment should be handled, e.g. telepresence or prepared-image
	Strategy string `json:"strategy,omitempty"`
	// Additional arguments to the given strategy
	Args map[string]string `json:"args,omitempty"`
}

// Route defines the strategy for how the traffic is routed to the Ref.
// +k8s:openapi-gen=true
type Route struct {
	// The type of route to use, e.g. header
	Type string `json:"type,omitempty"`
	// Name of the key, e.g. http header
	Name string `json:"name,omitempty"`
	// The value to use for routing
	Value string `json:"value,omitempty"`
}

func (r *Route) String() string {
	return fmt.Sprintf("%s:%s=%s", r.Type, r.Name, r.Value)
}

// SessionStatus defines the observed state of Session.
// +k8s:openapi-gen=true
type SessionStatus struct {
	State *SessionState `json:"state,omitempty"`
	// The current configured route
	Route *Route `json:"route,omitempty"`
	// The combined log of changes across all refs
	Conditions []*Condition `json:"conditions,omitempty"`

	Hosts []string `json:"hosts,omitempty"`

	// Fields below are solely for UX when inspecting CRDs from CLI, as the `additionalPrinterColumns` support only simple JSONPath expressions right now
	// See discussion on https://github.com/kubernetes/kubectl/issues/517 and linked issues about the limitation and status of the work

	// RouteExpression represents the Route object as single string expression
	RouteExpression string   `json:"_routeExp,omitempty"`   //nolint:tagliatelle //reason intentionally prefixed with _ to distinguish as UI/CLI field.
	Strategies      []string `json:"_strategies,omitempty"` //nolint:tagliatelle //reason intentionally prefixed with _ to distinguish as UI/CLI field.
	RefNames        []string `json:"_refNames,omitempty"`   //nolint:tagliatelle //reason intentionally prefixed with _ to distinguish as UI/CLI field.
}

// Condition .... .
// +k8s:openapi-gen=true
type Condition struct {
	Target Target `json:"target"`
	// Message explains the reason for a change.
	Message *string `json:"message,omitempty"`
	// Reason is a programmatic reason for the change.
	Reason *string `json:"reason,omitempty"`
	// Status indicates success.
	Status *string `json:"status,omitempty"`
	// Type the type of change
	Type *string `json:"type,omitempty"`
	// LastTransitionTime is the last time this action was applied
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

type Target struct {
	Name string `json:"name"`
	Ref  string `json:"ref"`
	Kind string `json:"kind"`
}

// +genclient
// +genclient:noStatus
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Ref Names",type="string",JSONPath=".status._refNames",description="refs being manipulated by this session"
// +kubebuilder:printcolumn:name="Strategies",type="string",JSONPath=".status._strategies",description="strategies used by session"
// +kubebuilder:printcolumn:name="Hosts",type="string",JSONPath=".status.hosts",description="exposed hosts for this session"
// +kubebuilder:printcolumn:name="Route",type="string",JSONPath=".status._routeExp",description="route expression used for this session"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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

// AddCondition adds or replaces a condition based on Name, Kind and Ref as a key.
func (s *Session) AddCondition(condition Condition) {
	replaced := false

	if condition.LastTransitionTime.IsZero() {
		now := metav1.NewTime(time.Now())
		condition.LastTransitionTime = &now
	}
	for i, stored := range s.Status.Conditions {
		if stored.Target.Name == condition.Target.Name &&
			stored.Target.Kind == condition.Target.Kind &&
			stored.Target.Ref == condition.Target.Ref {
			s.Status.Conditions[i] = &condition
			replaced = true
		}
	}
	if !replaced {
		s.Status.Conditions = append(s.Status.Conditions, &condition)
	}
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
