package model

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// StrategyExisting holds the name of the existing strategy
	StrategyExisting = "existing"
)

// SessionContext holds the context for a single session object, giving access to key things like REST Client and target Namespace.
type SessionContext struct {
	context.Context

	Name      string
	Namespace string
	UID       types.UID
	Route     Route
	Client    client.Client
	Log       logr.Logger
}

// ToOwnerReference returns a OwnerReference object that represents this Session
func (s *SessionContext) ToOwnerReference() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: "maistra.io/v1alpha1",
		Kind:       "Session",
		Name:       s.Name,
		UID:        s.UID,
	}
}

// Route references the strategy used to route to the target Refs.
type Route struct {
	Type  string
	Name  string
	Value string
}

// Ref references to a single Reference, e.g. Deployment, DeploymentConfig or GitRepo.
type Ref struct {
	Name             string
	Namespace        string
	Strategy         string
	Args             map[string]string
	Targets          []LocatedResourceStatus
	ResourceStatuses []ResourceStatus
}

// HostName represents the Hostname of a service in a given namespace.
type HostName struct {
	Name      string
	Namespace string
}

// Predicate base function to filter Resources.
type Predicate func(ResourceStatus) bool

// Any Predicate returns true if any of the predicates match.
func Any(predicates ...Predicate) Predicate {
	return func(resource ResourceStatus) bool {
		for _, predicate := range predicates {
			if predicate(resource) {
				return true
			}
		}
		return false
	}
}

// All Predicate returns true if all of the predicates match.
func All(predicates ...Predicate) Predicate {
	return func(resource ResourceStatus) bool {
		for _, predicate := range predicates {
			if !predicate(resource) {
				return false
			}
		}
		return true
	}
}

// Kind Predicate returns true if kind matches resource.
func Kind(kind string) Predicate {
	return func(resource ResourceStatus) bool {
		return resource.Kind == kind
	}
}

// Name Predicate returns true if name matches resource.
func Name(name string) Predicate {
	return func(resource ResourceStatus) bool {
		return resource.Name == name
	}
}

// AnyKind is a shortcut Predicate for Any and Kind from strings.
func AnyKind(kinds ...string) Predicate {
	pred := make([]Predicate, 0, len(kinds))
	for _, kind := range kinds {
		pred = append(pred, Kind(kind))
	}
	return Any(pred...)
}

// GetTargets use a Predicate to filter the LocatedResourceStatus.
func (r *Ref) GetTargets(predicate Predicate) []LocatedResourceStatus {
	var targets []LocatedResourceStatus
	for _, target := range r.Targets {
		if predicate(target.ResourceStatus) {
			targets = append(targets, target)
		}
	}
	return targets
}

// Match returns true if this Hostname is equal to the short or long v of a dns name.
func (h *HostName) Match(name string) bool {
	equalsShortName := h.Name == name
	equalsFullDNSName := fmt.Sprint(h.Name, ".", h.Namespace, ".svc.cluster.local") == name
	return equalsShortName || equalsFullDNSName
}

// GetTargetHostNames returns a list of Host names that the target Deployment can be reached under.
func (r *Ref) GetTargetHostNames() []HostName {
	targets := r.GetTargets(Kind("Service"))
	hosts := make([]HostName, 0, len(targets))
	for _, service := range targets {
		hosts = append(hosts, HostName{Name: service.Name, Namespace: r.Namespace})
	}

	return hosts
}

// GetVersion returns the existing version name.
func (r *Ref) GetVersion() string {
	target := r.GetTargets(AnyKind("Deployment", "DeploymentConfig"))
	if len(target) == 1 {
		if val, ok := target[0].Labels["version"]; ok {
			return val
		}
	}
	return "unknown"
}

// GetNewVersion returns the new updated version name.
func (r *Ref) GetNewVersion(sessionName string) string {
	return r.GetVersion() + "-" + sessionName
}

// AddTargetResource adds the status of an involved Resource to this ref.
func (r *Ref) AddTargetResource(ref LocatedResourceStatus) {
	replaced := false
	for i, status := range r.Targets {
		if ref.Name == status.Name && ref.Kind == status.Kind {
			r.Targets[i] = ref
			replaced = true
		}
	}
	if !replaced {
		r.Targets = append(r.Targets, ref)
	}
}

// AddResourceStatus adds the status of an involved Resource to this ref.
func (r *Ref) AddResourceStatus(ref ResourceStatus) {
	replaced := false
	for i, status := range r.ResourceStatuses {
		if ref.Name == status.Name && ref.Kind == status.Kind {
			r.ResourceStatuses[i] = ref
			replaced = true
		}
	}
	if !replaced {
		r.ResourceStatuses = append(r.ResourceStatuses, ref)
	}
}

// RemoveResourceStatus removes the status of an involved Resource to this ref.
func (r *Ref) RemoveResourceStatus(ref ResourceStatus) {
	for i, status := range r.ResourceStatuses {
		if ref.Name == status.Name && ref.Kind == status.Kind {
			r.ResourceStatuses = append(r.ResourceStatuses[:i], r.ResourceStatuses[i+1:]...)
		}
	}
}

// GetResources use a Predicate to filter the ResourceStatus.
func (r *Ref) GetResources(predicate Predicate) []ResourceStatus {
	var refs []ResourceStatus
	for _, status := range r.ResourceStatuses {
		if predicate(status) {
			refs = append(refs, status)
		}
	}
	return refs
}

// ResourceStatus holds information about the resources created/changed to fulfill a Ref.
type ResourceStatus struct {
	Kind string
	Name string
	// created, mutated, failed
	Action ResourceAction
	Prop   map[string]string
}

// LocatedResourceStatus is a ResourceStatus with labels.
type LocatedResourceStatus struct {
	ResourceStatus

	Labels map[string]string
}

// NewLocatedResource is a simple helper to create LocatedResourceStatus.
func NewLocatedResource(kind, name string, labels map[string]string) LocatedResourceStatus {
	return LocatedResourceStatus{
		ResourceStatus: ResourceStatus{
			Kind:   kind,
			Name:   name,
			Action: ActionLocated,
		},
		Labels: labels,
	}
}

// ResourceAction describes which type of operation was done/attempted to the target resource. Used to determine how to undo it.
type ResourceAction string

const (
	// ActionCreated imply the whole Named Kind was created and can be deleted
	ActionCreated ResourceAction = "created"
	// ActionModified imply the Named Kind has been modified and needs to be reverted to get back to original state
	ActionModified ResourceAction = "modified"
	// ActionFailed imply what ever was attempted failed. Assume current state is ok in clean up?
	ActionFailed ResourceAction = "failed"
	// ActionLocated imply the resource was found, but nothing was changed.
	ActionLocated ResourceAction = "located"
)

// Locator should attempt to resolve a Ref.Name to a target Named kind, e.g. Deployment, DeploymentConfig.
// return false if nothing was found.
type Locator func(SessionContext, *Ref) bool

// Mutator should create/modify a target Kind as required as needed
// * Add status to Ref for storage, failed or not.
type Mutator func(SessionContext, *Ref) error

// Revertor should delete/modify a target Kind to return to the original state after a Mutator
// * Remove status from Ref unless there is a failure that requires retry.
type Revertor func(SessionContext, *Ref) error
