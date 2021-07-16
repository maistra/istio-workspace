package model

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ModificatorStatusReporter func(ModificatorStatus)

type ModificatorController func(LocatorStatusStore) bool

// ModificatorStatus is the status of the given Locator action after it's been attempted performed.
type ModificatorStatus struct {
	LocatorStatus
	Error   error
	Success bool
	Prop    map[string]string
	Target  *Resource
}

type ModificatorRegistrar func() (targetResourceType client.Object, modificator Modificator)

// Modificator should perform the provided action on the given Resources provided by the Locator, e.g. Modify or Revert.
type Modificator func(context SessionContext, ref Ref, store LocatorStatusStore, reporter ModificatorStatusReporter)

// ModificatorStore Dummy impl for testing purposes.
type ModificatorStore struct {
	Stored []ModificatorStatus
}

func (m *ModificatorStore) Report(status ModificatorStatus) {
	m.Stored = append(m.Stored, status)
}
