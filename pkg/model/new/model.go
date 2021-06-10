package new

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

/*
	 Locator
	* Mutate
	* Revert

 	Mutator/Revertor

	ComponentLogger
	ConditionLogger
	EventLogger

	Accumulation Status




Session
	Ref
		- Y

	Status
		- Y



SessionController -> Convert(Session->Model) -> Engine -> Locator|Mutator -> ConditionLogger(Convert(Model->Session))
                                                                          -> ComponentLogger(Convert(Model->Session))
                                                                          -> EventLogger(Convert(Model->Event API))


Locator <-- Object | List
	Service
	Gateway
	VitualService
	DestinationRule
	Deployment
	DeploymentConfig
	...

Modificator <-- Object | List
	Gateway
	VitualService
	DestinationRule
	Deployment
	DeploymentConfig

*/

/*
 *  MODEL API - Level
 */

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

type Ref struct {
	KindName  RefKindName
	Deleted   bool
	Namespace string
	Strategy  string
	Args      map[string]string
}

type RefKindName struct {
	Kind string
	Name string
}

// HostName represents the Hostname of a service in a given namespace.
type HostName struct {
	Name      string
	Namespace string
}

type LocatorStatus struct {
	Namespace string
	Kind      string
	Name      string
	TimeStamp time.Time
	Labels    map[string]string
	Action    StatusAction // Create, Modify, Located
}

type LocatorStatusReporter func(LocatorStatus)
type LocatedReporter func(LocatorStatusStore)
type LocatorStatusStore func(kind ...string) []LocatorStatus

type Locator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter LocatorStatusReporter) error

type ModificatorStatusReporter func(ModificatorStatus)

type ModificatorController func(LocatorStatusStore) bool

type ModificatorStatus struct {
	LocatorStatus
	Error   error
	Success bool
	Prop    map[string]string
}

type ModificatorRegistrar func() (targetResourceType client.Object, modificator Modificator)
type Modificator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter ModificatorStatusReporter)

type Sync func(SessionContext, Ref, ModificatorController, LocatedReporter, ModificatorStatusReporter)

// TODO dummy impl for testing purposes.
type LocatorStore struct {
	stored []LocatorStatus
}

func (l *LocatorStore) Store(kind ...string) []LocatorStatus {
	if len(kind) == 0 {
		return l.stored
	}
	var f []LocatorStatus
	for _, loc := range l.stored {
		for _, k := range kind {
			if loc.Kind == k {
				f = append(f, loc)

				break
			}
		}
	}

	return f
}

func (l *LocatorStore) Report(status LocatorStatus) {
	replaced := false

	if status.TimeStamp.IsZero() {
		status.TimeStamp = time.Now()
	}
	for i, stored := range l.stored {
		if stored.Name == status.Name && stored.Kind == status.Kind && stored.Action == status.Action {
			l.stored[i] = status
			replaced = true
		}
	}
	if !replaced {
		l.stored = append(l.stored, status)
	}
}

func (l *LocatorStore) Clear() {
	l.stored = []LocatorStatus{}
}

// ModificatorStore Dummy impl for testing purposes.
type ModificatorStore struct {
	Stored []ModificatorStatus
}

func (m *ModificatorStore) Report(status ModificatorStatus) {
	m.Stored = append(m.Stored, status)
}

// String returns the string formatted kind/name.
func (r RefKindName) String() string {
	if r.Kind == "" {
		return r.Name
	}

	return r.Kind + "/" + r.Name
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

// SupportsKind returns true if kind match or the kind is empty.
func (r RefKindName) SupportsKind(kind string) bool {
	return r.Kind == "" || strings.EqualFold(r.Kind, kind)
}

// GetVersion returns the existing version name.
func GetVersion(store LocatorStatusStore) string {
	target := store("Deployment", "DeploymentConfig")
	if len(target) == 1 {
		if val, ok := target[0].Labels["version"]; ok {
			return val
		}
	}

	return "unknown"
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

func ParseHostName(host string) HostName {
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

// GetNewVersion returns the new updated version name.
func GetNewVersion(store LocatorStatusStore, sessionName string) string {
	return GetSha(GetVersion(store)) + "-" + sessionName
}

// GetSha computes a hash of the version and returns 8 characters substring of it.
func GetSha(version string) string {
	sum := sha256.Sum256([]byte(version))
	sha := fmt.Sprintf("%x", sum)

	return sha[:8]
}

/*
 *  Session Controller - Level
 */

type Validator func([]LocatorStatus) error

type Condition struct {
	// Human readable reason for the change
	Message string `json:"message,omitempty"`
	// Programmatic reason for the change
	Reason string `json:"reason,omitempty"`
	// Boolean value to indicate success
	Status string `json:"status,omitempty"`
	// The type of change
	Type string `json:"type,omitempty"`
}

/*
 *  Impl
 */

func EngineImpl(locators []Locator, modificators []Modificator) Sync {
	return func(context SessionContext, ref Ref, modify ModificatorController, locatedReporter LocatedReporter, modificationReporter ModificatorStatusReporter) {
		located := LocatorStore{}
		for _, locator := range locators {
			err := locator(
				context,
				ref,
				located.Store,
				located.Report,
			)

			if err != nil {
				// TODO what do we do here?
			}
		}
		if !modify(located.Store) {
			return
		}

		locatedReporter(located.Store)

		for _, modificator := range modificators {
			modificator(context, ref, located.Store, modificationReporter)
		}
	}
}
