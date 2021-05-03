package model

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// StrategyExisting holds the name of the existing strategy.
	StrategyExisting = "existing"
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

// ToNamespacedName returns a types.NamespaceName object that represents this Session.
func (s *SessionContext) ToNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Namespace: s.Namespace,
		Name:      s.Name,
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
	KindName         RefKindName
	Namespace        string
	Strategy         string
	Args             map[string]string
	Targets          []LocatedResourceStatus
	ResourceStatuses []ResourceStatus
}

// RefKindName holds the optionally specified Kind together with the name, e.g. deploymentconfig/name.
type RefKindName struct {
	Kind string
	Name string
}

// String returns the string formatted kind/name.
func (r RefKindName) String() string {
	if r.Kind == "" {
		return r.Name
	}

	return r.Kind + "/" + r.Name
}

// From parses a String() representation into a Object.
func ParseRefKindName(exp string) RefKindName {
	trimmedExp := strings.TrimSpace(strings.ToLower(exp))
	parts := strings.Split(trimmedExp, "/")
	if len(parts) == 2 {
		return RefKindName{
			Kind: parts[0],
			Name: parts[1],
		}
	}

	return RefKindName{Name: trimmedExp}
}

// SupportsKind returns true if kind match or the kind is empty.
func (r RefKindName) SupportsKind(kind string) bool {
	return r.Kind == "" || strings.EqualFold(r.Kind, kind)
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

// String returns the String representation of a HostName
func (h *HostName) String() string {
	if h.Namespace != "" {
		return fmt.Sprint(h.Name, ".", h.Namespace, ".svc.cluster.local")
	}
	return h.Name
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
	return GetSha(r.GetVersion()) + "-" + sessionName
}

// AddTargetResource adds the status of an involved Resource to this ref.
func (r *Ref) AddTargetResource(ref LocatedResourceStatus) {
	replaced := false

	if ref.TimeStamp.IsZero() {
		ref.TimeStamp = time.Now()
	}
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

	if ref.TimeStamp.IsZero() {
		ref.TimeStamp = time.Now()
	}
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

// GetSha computes a hash of the version and returns 8 characters substring of it.
func GetSha(version string) string {
	sum := sha256.Sum256([]byte(version))
	sha := fmt.Sprintf("%x", sum)

	return sha[:8]
}

// ResourceStatus holds information about the resources created/changed to fulfill a Ref.
type ResourceStatus struct {
	Kind      string
	Name      string
	TimeStamp time.Time
	Action    ResourceAction
	Success   bool
	Message   string
	Prop      map[string]string
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
			Kind:    kind,
			Name:    name,
			Action:  ActionLocated,
			Success: true,
		},
		Labels: labels,
	}
}

// NewFailedResource is a simple helper to create ResourceStatus with failed status.
func NewFailedResource(kind, name string, action ResourceAction, message string) ResourceStatus {
	return ResourceStatus{
		Kind:    kind,
		Name:    name,
		Action:  action,
		Success: false,
		Message: message,
	}
}

// NewSuccessResource is a simple helper to create ResourceStatus with success status.
func NewSuccessResource(kind, name string, action ResourceAction) ResourceStatus {
	return ResourceStatus{
		Kind:    kind,
		Name:    name,
		Action:  action,
		Success: true,
	}
}

// ResourceAction describes which type of operation was done/attempted to the target resource. Used to determine how to undo it.
type ResourceAction string

const (
	// ActionCreated imply the whole Named Kind was created and can be deleted.
	ActionCreated ResourceAction = "created"
	// ActionModified imply the Named Kind has been modified and needs to be reverted to get back to original state.
	ActionModified ResourceAction = "modified"
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

// Manipulator is a joint interface between a Mutator and Revertor function combined with their target Object.
type Manipulator interface {
	// Mutate is called to manipulate/mutate the TargetResource
	Mutate() Mutator

	// Revert is called to undo the manipulated/mutated TargetResource
	Revert() Revertor

	// TargetResourceType should return a empty version of the Type of Resource this manipulator targets.
	// Used to register Watch for changes to the Type.
	TargetResourceType() client.Object
}
