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
	Client    client.Client
	Log       logr.Logger
}

// Ref refereces to a single Reference, e.g. Deployment, DeploymentConfig or GitRepo
type Ref struct {
	Name             string
	ResourceStatuses []ResourceStatus
}

// AddResourceStatus adds the status of an involved Resource to this ref
func (r *Ref) AddResourceStatus(ref ResourceStatus) {
	r.ResourceStatuses = append(r.ResourceStatuses, ref)
}

// GetResourceStatus returns a array of involved Resources based on a k8 Kind
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
	// Created, Mutated
	Action ResourceAction
}

// ResourceAction describes which type of operation was done/attempted to the target resource. Used to determine how to undo it.
type ResourceAction int

const (
	// ActionCreated imply the whole Named Kind was created and can be deleted
	ActionCreated ResourceAction = iota
	// ActionModified imply the Named Kind has been modified and needs to be reverted to get back to original state
	ActionModified ResourceAction = iota
	// ActionFailed imply what ever was attempted failed. Assume current state is ok in clean up?
	ActionFailed ResourceAction = iota
)

// Locator should attempt to resolve a Ref.Name to a target Named kind, e.g. Deployment, DeploymentConfig.
// return false if nothing was found
type Locator func(SessionContext, *Ref) bool

// Mutator should create/modify a target Kind as required as needed
// * Add status to Ref for storage
type Mutator func(SessionContext, *Ref) error

// Revertor should delete/modify a target Kind to return to the original state after a Mutator
// * Don't add status to Ref unless failure that requires retry
type Revertor func(SessionContext, *Ref) error
