package model

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SessionContext holds the context for a single session object, giving access to key things like REST Client and target Namespace.
type SessionContext struct {
	context.Context

	Name      string
	Namespace string
	Route     Route
	Client    client.Client
	Log       logr.Logger
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

// GetTargetsByKind returns the targets of the given kind.
func (r *Ref) GetTargetsByKind(kinds ...string) []LocatedResourceStatus {
	targets := []LocatedResourceStatus{}
	for _, target := range r.Targets {
		for _, kind := range kinds {
			if target.Kind == kind {
				targets = append(targets, target)
				break
			}
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
	hosts := []HostName{}
	for _, service := range r.GetTargetsByKind("Service") {
		hosts = append(hosts, HostName{Name: service.Name, Namespace: r.Namespace})
	}

	return hosts
}

// GetVersion returns the existing version name.
func (r *Ref) GetVersion() string {
	target := r.GetTargetsByKind("Deployment", "DeploymentConfig")
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

// GetResourceStatus returns a array of involved Resources based on a k8s Kind.
func (r *Ref) GetResourceStatus(kind string) []ResourceStatus {
	refs := []ResourceStatus{}
	for _, status := range r.ResourceStatuses {
		if status.Kind == kind {
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
