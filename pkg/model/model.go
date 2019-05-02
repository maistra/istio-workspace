package model

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SessionContext holds the context for a single session object, giving access to key things like REST Client and target Namespace
type SessionContext struct {
	context.Context

	Name      string
	Namespace string
	Route     Route
	Client    client.Client
	Log       logr.Logger
}

// Route references the strategy used to route to the target Refs
type Route struct {
	Type  string
	Name  string
	Value string
}

// Ref references to a single Reference, e.g. Deployment, DeploymentConfig or GitRepo
type Ref struct {
	Name             string
	ResourceStatuses []ResourceStatus
}

// AddResourceStatus adds the status of an involved Resource to this ref
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

// RemoveResourceStatus removes the status of an involved Resource to this ref
func (r *Ref) RemoveResourceStatus(ref ResourceStatus) {
	for i, status := range r.ResourceStatuses {
		if ref.Name == status.Name && ref.Kind == status.Kind {
			r.ResourceStatuses = append(r.ResourceStatuses[:i], r.ResourceStatuses[i+1:]...)
		}
	}
}

// GetResourceStatus returns a array of involved Resources based on a k8s Kind
func (r *Ref) GetResourceStatus(kind string) []ResourceStatus {
	refs := []ResourceStatus{}
	for _, status := range r.ResourceStatuses {
		if status.Kind == kind {
			refs = append(refs, status)
		}
	}
	return refs
}

// ResourceStatus holds information about the resources created/changed to fulfill a Ref
type ResourceStatus struct {
	Kind string
	Name string
	// created, mutated, failed
	Action ResourceAction
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
)

// Locator should attempt to resolve a Ref.Name to a target Named kind, e.g. Deployment, DeploymentConfig.
// return false if nothing was found
type Locator func(SessionContext, *Ref) bool

// Mutator should create/modify a target Kind as required as needed
// * Add status to Ref for storage, failed or not
type Mutator func(SessionContext, *Ref) error

// Revertor should delete/modify a target Kind to return to the original state after a Mutator
// * Remove status from Ref unless there is a failure that requires retry
type Revertor func(SessionContext, *Ref) error
