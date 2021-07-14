package model

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ModificatorStatusReporter func(ModificatorStatus)

type ModificatorController func(LocatorStatusStore) bool

type ModificatorStatus struct {
	LocatorStatus
	Error   error
	Success bool
	Prop    map[string]string
	Target  *Resource
}

type ModificatorRegistrar func() (targetResourceType client.Object, modificator Modificator)
type Modificator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter ModificatorStatusReporter)

type Sync func(SessionContext, Ref, ModificatorController, LocatedReporter, ModificatorStatusReporter)

// ModificatorStore Dummy impl for testing purposes.
type ModificatorStore struct {
	Stored []ModificatorStatus
}

func (m *ModificatorStore) Report(status ModificatorStatus) {
	m.Stored = append(m.Stored, status)
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
