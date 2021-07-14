package model

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// StatusAction describes which type of operation was done/attempted to the target resource. Used to determine how to undo it.
type StatusAction string

const (
	// ActionCreate imply the whole Named Kind should be created.
	ActionCreate StatusAction = "create"
	// ActionDelete imply the whole Named Kind was created and should be deleted.
	ActionDelete StatusAction = "delete"
	// ActionModify imply the Named Kind should be modified.
	ActionModify StatusAction = "modify"
	// ActionRevert imply the Named Kind was modified and should be reverted to original state.
	ActionRevert StatusAction = "revert"
	// ActionLocated imply the resource was found, but nothing was changed.
	ActionLocated StatusAction = "located"

	// StrategyExisting holds the name of the existing strategy.
	StrategyExisting = "existing"
)

func Flip(action StatusAction) StatusAction {
	switch action {
	case ActionCreate:
		return ActionDelete
	case ActionDelete:
		return ActionCreate
	case ActionModify:
		return ActionRevert
	case ActionRevert:
		return ActionModify
	case ActionLocated:
		return ActionLocated
	}

	return ActionRevert
}

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

// Ref references the user specified Resource target and configuration.
type Ref struct {
	KindName  RefKindName
	Deleted   bool // TODO rename to something more indicating the intent vs state (e.g. MarkedForDeletion)
	Namespace string
	Strategy  string
	Args      map[string]string
}

// Hash returns a predictable hash version for this object.
func (r *Ref) Hash() string {
	digest := "kind:" + r.KindName.String()
	digest += ";deleted:" + strconv.FormatBool(r.Deleted)
	digest += ";namespace:" + r.Namespace
	digest += ";strategy:" + r.Strategy

	args := []string{}
	for k := range r.Args {
		args = append(args, k)
	}
	sort.Strings(args)

	for _, k := range args {
		digest += ";args[" + k + "]:" + r.Args[k]
	}

	sum := sha256.Sum256([]byte(digest))
	sha := fmt.Sprintf("%x", sum)

	return sha[:8]
}

// RefKindName is the target Resource and optional Resource Kind.
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

// SupportsKind returns true if kind match or the kind is empty.
func (r RefKindName) SupportsKind(kind string) bool {
	return r.Kind == "" || strings.EqualFold(r.Kind, kind)
}

// ParseRefKindName parses a String() representation into a Object.
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

// HostName represents the Hostname of a service in a given namespace.
type HostName struct {
	Name      string
	Namespace string
}

// Match returns true if this Hostname is equal to the short or long v of a dns name.
func (h *HostName) Match(name string) bool {
	equalsShortName := h.Name == name
	equalsFullDNSName := fmt.Sprint(h.Name, ".", h.Namespace, ".svc.cluster.local") == name

	return equalsShortName || equalsFullDNSName
}

// String returns the String representation of a HostName.
func (h *HostName) String() string {
	if h.Namespace != "" {
		return fmt.Sprint(h.Name, ".", h.Namespace, ".svc.cluster.local")
	}

	return h.Name
}

func NewHostName(host string) HostName {
	if strings.Contains(host, ".svc.cluster.local") {
		parts := strings.Split(host, ".")

		return HostName{Name: parts[0], Namespace: parts[1]}
	}

	return HostName{Name: host}
}

// GetTargetHostNames returns a list of Host names that the target Deployment can be reached under.
func GetTargetHostNames(store LocatorStatusStore) []HostName {
	targets := store("Service")
	hosts := make([]HostName, 0, len(targets))
	for _, service := range targets {
		hosts = append(hosts, HostName{Name: service.Name, Namespace: service.Namespace})
	}

	return hosts
}

type Resource struct {
	Namespace string
	Kind      string
	Name      string
}
